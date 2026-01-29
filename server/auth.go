package server

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthManager struct {
	DB       *sql.DB
	Sessions map[string]int64 // Token -> Expiry
	mu       sync.RWMutex
}

func NewAuthManager(gormDB *gorm.DB) *AuthManager {
	db, err := gormDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Table
	query := `
	CREATE TABLE IF NOT EXISTS auth (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		password_hash TEXT NOT NULL
	);
	`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	return &AuthManager{
		DB:       db,
		Sessions: make(map[string]int64),
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
		return false // No password set or error
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (am *AuthManager) CreateSession() string {
	token := uuid.New().String()
	am.mu.Lock()
	defer am.mu.Unlock()
	// 24 hour session
	am.Sessions[token] = time.Now().Add(24 * time.Hour).Unix()
	return token
}

func (am *AuthManager) ValidateSession(token string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	expiry, exists := am.Sessions[token]
	if !exists {
		return false
	}
	if time.Now().Unix() > expiry {
		delete(am.Sessions, token)
		return false
	}
	return true
}

func (am *AuthManager) RevokeSession(token string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.Sessions, token)
}

// Middleware
func (am *AuthManager) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Always allow static resources if they are not API routes
		// But here we are protecting the API specifically
		if !isApiRoute(path) {
			return c.Next()
		}

		// Public API routes
		if path == "/api/auth/status" || path == "/api/auth/setup" || path == "/api/auth/login" {
			return c.Next()
		}

		// Check Setup
		if !am.IsSetup() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Setup required",
				"code":  "SETUP_REQUIRED",
			})
		}

		// Check Token
		token := c.Cookies("auth_token")
		if token == "" || !am.ValidateSession(token) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		return c.Next()
	}
}

func isApiRoute(path string) bool {
	return len(path) >= 4 && path[:4] == "/api"
}
