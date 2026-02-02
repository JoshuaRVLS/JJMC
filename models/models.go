package models

type Instance struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Directory    string `json:"directory"`
	Type         string `json:"type"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	MaxMemory    int    `json:"maxMemory"`
	JavaArgs     string `json:"javaArgs"`
	JarFile      string `json:"jarFile"`
	JavaPath     string `json:"javaPath"`
	StartCommand string `json:"startCommand"`
	WebhookURL   string `json:"webhookUrl"`
}

type InstanceModel struct {
	ID           string `gorm:"primaryKey"`
	Name         string
	Type         string
	Version      string
	MaxMemory    int
	JavaArgs     string
	JarFile      string
	JavaPath     string
	StartCommand string
	WebhookURL   string
	CreatedAt    int64
}
