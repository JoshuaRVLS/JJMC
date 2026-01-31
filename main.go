package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"jjmc/auth"
	"jjmc/database"
	"jjmc/instances"
	"jjmc/pkg/logger"
	"jjmc/pkg/signals"
	"jjmc/services"
	"jjmc/web"

	"jjmc/rcon"
	"jjmc/sftp"
	"jjmc/telnet"

	"github.com/gofiber/fiber/v2"
)

func main() {
	silent := flag.Bool("silent", false, "Suppress server logs in terminal")
	flag.Parse()

	logger.Setup()

	database.ConnectDB()

	authManager := auth.NewAuthManager(database.DB)

	templateManager := services.NewTemplateManager("./templates")
	if err := templateManager.LoadTemplates(); err != nil {
		logger.Warn("Failed to load templates", "error", err)
	}

	instanceManager := instances.NewInstanceManager(
		"./servers",
		templateManager,
		*silent,
	)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	web.RegisterRoutes(app, authManager, instanceManager)

	app.Static("/_app", "./build/_app", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 365 * 24 * time.Hour,
		MaxAge:        31536000,
	})

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

	app.Get("*", func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.SendFile("./build/index.html")
	})

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

	sftpServer := sftp.NewSFTPServer("0.0.0.0:2022", "./instances", authManager)
	go func() {
		if err := sftpServer.Start(); err != nil {
			logger.Error("SFTP Server failed to start", "error", err)
		}
	}()

	telnetServer := telnet.NewTelnetServer("0.0.0.0:2023", authManager, instanceManager)
	go func() {
		if err := telnetServer.Start(); err != nil {
			logger.Error("Telnet Server failed to start", "error", err)
		}
	}()

	rconServer := rcon.NewRCONServer("0.0.0.0:2024", authManager, instanceManager)
	go func() {
		if err := rconServer.Start(); err != nil {
			logger.Error("RCON Server failed to start", "error", err)
		}
	}()

	go func() {
		if err := app.Listen("0.0.0.0:3000"); err != nil {
			logger.Error("Web Server failed to start", "error", err)
		}
	}()

	ctx := signals.SetupSignalHandler()
	<-ctx.Done()

	logger.Info("Shutting down servers...")

	if err := app.Shutdown(); err != nil {
		logger.Error("Error shutting down web server", "error", err)
	}

	sftpServer.Close()
	telnetServer.Close()
	rconServer.Close()

	logger.Info("Shutdown complete.")
}
