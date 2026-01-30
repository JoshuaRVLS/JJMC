package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"jjmc/auth"
	"jjmc/database"
	"jjmc/instances"
	"jjmc/services"
	"jjmc/web"

	"jjmc/rcon"
	"jjmc/sftp"
	"jjmc/telnet"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize Database
	database.ConnectDB()

	// Initialize Managers
	authManager := auth.NewAuthManager(database.DB)

	// Template Manager
	templateManager := services.NewTemplateManager("./templates")
	if err := templateManager.LoadTemplates(); err != nil {
		fmt.Printf("Warning: Failed to load templates: %v\n", err)
	}

	// Initialize Instances
	instanceManager := instances.NewInstanceManager(
		"./servers",
		templateManager,
	)

	// Setup Fiber
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Register API Routes
	web.RegisterRoutes(app, authManager, instanceManager)

	// Serve Immutable Assets (Long Cache)
	app.Static("/_app", "./build/_app", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 365 * 24 * time.Hour,
		MaxAge:        31536000,
	})

	// Serve Root/Index (No Cache)
	app.Static("/", "./build", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 0,
		MaxAge:        0,
		ModifyResponse: func(c *fiber.Ctx) error {
			c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
			return nil
		},
	})

	// SPA Fallback
	app.Get("*", func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.SendFile("./build/index.html")
	})

	// Log Addresses
	if ifaces, err := net.Interfaces(); err == nil {
		fmt.Println("Server available at:")
		fmt.Println("  http://localhost:3000")

		for _, i := range ifaces {
			if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
				continue
			}
			name := strings.ToLower(i.Name)
			if strings.Contains(name, "docker") || strings.Contains(name, "veth") ||
				strings.Contains(name, "br-") || strings.Contains(name, "virbr") ||
				strings.Contains(name, "vmnet") || strings.Contains(name, "wsl") {
				continue
			}

			addrs, err := i.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() {
					continue
				}
				if ip.To4() != nil {
					fmt.Printf("  http://%s:3000\n", ip.String())
				}
			}
		}
	}

	// Start SFTP Server
	go func() {
		sftpServer := sftp.NewSFTPServer("0.0.0.0:2022", "./instances", authManager)
		if err := sftpServer.Start(); err != nil {
			log.Printf("SFTP Server failed to start: %v", err)
		}
	}()

	// Start Telnet Server
	go func() {
		telnetServer := telnet.NewTelnetServer("0.0.0.0:2023", authManager, instanceManager)
		if err := telnetServer.Start(); err != nil {
			log.Printf("Telnet Server failed to start: %v", err)
		}
	}()

	// Start RCON Server
	go func() {
		rconServer := rcon.NewRCONServer("0.0.0.0:2024", authManager, instanceManager)
		if err := rconServer.Start(); err != nil {
			log.Printf("RCON Server failed to start: %v", err)
		}
	}()

	log.Fatal(app.Listen("0.0.0.0:3000"))
}
