package models

type Instance struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Directory    string `json:"directory"`
	Type         string `json:"type"` // e.g. "fabric", "vanilla"
	Version      string `json:"version"`
	Status       string `json:"status"` // "Online" or "Offline"
	MaxMemory    int    `json:"maxMemory"`
	JavaArgs     string `json:"javaArgs"`
	JarFile      string `json:"jarFile"`
	StartCommand string `json:"startCommand"`
}

type InstanceModel struct {
	ID           string `gorm:"primaryKey"`
	Name         string
	Type         string
	Version      string
	MaxMemory    int
	JavaArgs     string
	JarFile      string
	StartCommand string
	CreatedAt    int64
}
