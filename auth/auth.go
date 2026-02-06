package auth

import (
	"database/sql"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthManager struct {
	DB       *sql.DB
	LaunchID string
	mu       sync.RWMutex
}

func NewAuthManager(gormDB *gorm.DB) *AuthManager {
	db, err := gormDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS auth (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			password_hash TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			expiry INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS login_attempts (
			ip TEXT PRIMARY KEY,
			attempts INTEGER NOT NULL,
			last_attempt INTEGER NOT NULL
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Fatal(err)
		}
	}

	// Clean up expired sessions on startup
	db.Exec("DELETE FROM sessions WHERE expiry < ?", time.Now().Unix())

	return &AuthManager{
		DB:       db,
		LaunchID: uuid.New().String(),
	}
}

func (am *AuthManager) IsSetup() bool {
	var count int
	err := am.DB.QueryRow("SELECT COUNT(*) FROM auth WHERE id = 1").Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func (am *AuthManager) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = am.DB.Exec("INSERT OR REPLACE INTO auth (id, password_hash) VALUES (1, ?)", string(hash))
	return err
}

func (am *AuthManager) VerifyPassword(password string) bool {
	var hash string
	err := am.DB.QueryRow("SELECT password_hash FROM auth WHERE id = 1").Scan(&hash)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (am *AuthManager) CreateSession() string {
	token := uuid.New().String()
	expiry := time.Now().Add(24 * time.Hour).Unix()

	_, err := am.DB.Exec("INSERT INTO sessions (token, expiry) VALUES (?, ?)", token, expiry)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return ""
	}
	return token
}

func (am *AuthManager) ValidateSession(token string) bool {
	var expiry int64
	err := am.DB.QueryRow("SELECT expiry FROM sessions WHERE token = ?", token).Scan(&expiry)
	if err != nil {
		return false
	}

	if time.Now().Unix() > expiry {
		am.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
		return false
	}
	return true
}

func (am *AuthManager) RevokeSession(token string) {
	am.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
}

// Rate Limiting
const (
	MaxAttempts     = 5
	LockoutDuration = 15 * time.Minute // 15 minutes
	AttemptWindow   = 5 * time.Minute  // Reset attempts if no activity for 5 minutes
)

func (am *AuthManager) CheckRateLimit(ip string) (bool, time.Duration) {
	am.mu.Lock()
	defer am.mu.Unlock()

	var attempts int
	var lastAttempt int64
	err := am.DB.QueryRow("SELECT attempts, last_attempt FROM login_attempts WHERE ip = ?", ip).Scan(&attempts, &lastAttempt)

	if err == sql.ErrNoRows {
		return true, 0
	} else if err != nil {
		log.Printf("Rate limit check error: %v", err)
		return true, 0 // Fail open on DB error to avoid locking everyone out
	}

	last := time.Unix(lastAttempt, 0)

	// If currently locked out
	if attempts >= MaxAttempts {
		lockoutExpire := last.Add(LockoutDuration)
		if time.Now().Before(lockoutExpire) {
			return false, time.Until(lockoutExpire)
		}
		// Lockout expired, reset
		am.DB.Exec("DELETE FROM login_attempts WHERE ip = ?", ip)
		return true, 0
	}

	// If attempt window expired, reset attempts (but we'll update it in RecordAttempt)
	if time.Since(last) > AttemptWindow {
		am.DB.Exec("DELETE FROM login_attempts WHERE ip = ?", ip)
		return true, 0
	}

	return true, 0
}

func (am *AuthManager) RecordLoginAttempt(ip string, success bool) {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now().Unix()

	if success {
		am.DB.Exec("DELETE FROM login_attempts WHERE ip = ?", ip)
		return
	}

	// Increment attempts
	_, err := am.DB.Exec(`
		INSERT INTO login_attempts (ip, attempts, last_attempt) 
		VALUES (?, 1, ?) 
		ON CONFLICT(ip) DO UPDATE SET 
			attempts = attempts + 1, 
			last_attempt = excluded.last_attempt
	`, ip, now)

	if err != nil {
		log.Printf("Failed to record login attempt: %v", err)
	}
}

func (am *AuthManager) GetRemainingAttempts(ip string) int {
	var attempts int
	err := am.DB.QueryRow("SELECT attempts FROM login_attempts WHERE ip = ?", ip).Scan(&attempts)
	if err != nil {
		return MaxAttempts
	}
	return int(math.Max(0, float64(MaxAttempts-attempts)))
}
