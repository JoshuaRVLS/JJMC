package server

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type InstanceModel struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	Type      string
	Version   string
	MaxMemory int
	JavaArgs  string
	CreatedAt int64
}

func ConnectDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("jjmc.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto Migrate
	DB.AutoMigrate(&InstanceModel{})
}
