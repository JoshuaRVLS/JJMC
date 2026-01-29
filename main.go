package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"jjmc/server"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type SystemFile struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

func main() {
	app := fiber.New()

	app.Use(cors.New())

	// Initialize Database
	server.ConnectDB()

	// Initialize Auth Manager (Shared DB)
	authManager := server.NewAuthManager(server.DB)

	// Middleware
	app.Use(authManager.Middleware())

	// Auth Routes
	app.Get("/api/auth/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"isSetup":       authManager.IsSetup(),
			"authenticated": authManager.ValidateSession(c.Cookies("auth_token")),
		})
	})

	app.Post("/api/auth/setup", func(c *fiber.Ctx) error {
		if authManager.IsSetup() {
			return c.Status(400).JSON(fiber.Map{"error": "Already setup"})
		}
		var payload struct {
			Password string `json:"password"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}
		if len(payload.Password) < 8 {
			return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
		}
		if err := authManager.SetPassword(payload.Password); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Auto login
		token := authManager.CreateSession()
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			SameSite: "Strict",
		})

		return c.JSON(fiber.Map{"status": "setup_complete"})
	})

	app.Post("/api/auth/login", func(c *fiber.Ctx) error {
		var payload struct {
			Password string `json:"password"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		if !authManager.VerifyPassword(payload.Password) {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid password"})
		}

		token := authManager.CreateSession()
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			SameSite: "Strict",
		})

		return c.JSON(fiber.Map{"status": "logged_in"})
	})

	app.Post("/api/auth/logout", func(c *fiber.Ctx) error {
		token := c.Cookies("auth_token")
		if token != "" {
			authManager.RevokeSession(token)
		}
		c.ClearCookie("auth_token")
		return c.JSON(fiber.Map{"status": "logged_out"})
	})

	// Initialize Instance Manager
	instanceManager := server.NewInstanceManager("./instances")
	// No global versions manager, we create one per instance or pass instance to it?
	// Actually VersionsManager just needs options to install.
	// We can keep a global helper for downloading/installing to a directory.

	// API Routes for Instances
	app.Get("/api/instances", func(c *fiber.Ctx) error {
		return c.JSON(instanceManager.ListInstances())
	})

	app.Get("/api/system/files", func(c *fiber.Ctx) error {
		dirPath := c.Query("path")
		if dirPath == "" {
			// Default to user home directory
			home, err := os.UserHomeDir()
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to get home directory"})
			}
			dirPath = home
		}

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to read directory: %v", err)})
		}

		var files []SystemFile
		// potential parent directory
		parent := filepath.Dir(dirPath)
		if parent != dirPath {
			files = append(files, SystemFile{
				Name:  "..",
				Path:  parent,
				IsDir: true,
			})
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			files = append(files, SystemFile{
				Name:  entry.Name(),
				Path:  filepath.Join(dirPath, entry.Name()),
				IsDir: entry.IsDir(),
			})
		}

		return c.JSON(fiber.Map{
			"path":  dirPath,
			"files": files,
		})
	})

	app.Get("/api/versions/game", func(c *fiber.Ctx) error {
		resp, err := http.Get("https://api.modrinth.com/v2/tag/game_version")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch versions"})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read versions"})
		}

		c.Set("Content-Type", "application/json")
		return c.Send(body)
	})

	app.Get("/api/versions/loader", func(c *fiber.Ctx) error {
		resp, err := http.Get("https://api.modrinth.com/v2/tag/loader")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch loaders"})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read loaders"})
		}

		c.Set("Content-Type", "application/json")
		return c.Send(body)
	})

	app.Post("/api/instances/import", func(c *fiber.Ctx) error {
		var payload struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			SourcePath string `json:"sourcePath"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}
		inst, err := instanceManager.ImportInstance(payload.ID, payload.Name, payload.SourcePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(inst)
	})

	app.Post("/api/instances", func(c *fiber.Ctx) error {
		var payload struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Type    string `json:"type"`
			Version string `json:"version"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}
		inst, err := instanceManager.CreateInstance(payload.ID, payload.Name, payload.Type, payload.Version)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(inst)
	})

	app.Delete("/api/instances/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := instanceManager.DeleteInstance(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "deleted"})
	})

	app.Get("/api/instances/:id", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		return c.JSON(inst)
	})

	app.Patch("/api/instances/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var payload struct {
			MaxMemory int    `json:"maxMemory"`
			JavaArgs  string `json:"javaArgs"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		if err := instanceManager.UpdateSettings(id, payload.MaxMemory, payload.JavaArgs); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "updated"})
	})

	// System Routes
	app.Get("/api/system/uuid", func(c *fiber.Ctx) error {
		name := c.Query("name")
		offline := c.Query("offline") == "true"

		if name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
		}

		if offline {
			// Generate Offline UUID (Version 3 MD5 of "OfflinePlayer:" + name)
			data := []byte("OfflinePlayer:" + name)
			hash := md5.Sum(data)
			hash[6] = (hash[6] & 0x0f) | 0x30 // Version 3
			hash[8] = (hash[8] & 0x3f) | 0x80 // Variant 10 (IETF)

			uuid := fmt.Sprintf("%x-%x-%x-%x-%x", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:])

			return c.JSON(fiber.Map{
				"uuid":     uuid,
				"username": name,
			})
		}

		resp, err := http.Get(fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", name))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to contact Mojang API"})
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		} else if resp.StatusCode != 200 {
			return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "Mojang API error"})
		}

		var mojangResp struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&mojangResp); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to parse Mojang response"})
		}

		// Insert dashes into UUID
		uuidRaw := mojangResp.ID
		var uuid string
		if len(uuidRaw) == 32 {
			uuid = fmt.Sprintf("%s-%s-%s-%s-%s", uuidRaw[0:8], uuidRaw[8:12], uuidRaw[12:16], uuidRaw[16:20], uuidRaw[20:])
		} else {
			uuid = uuidRaw // Should warn or error, but fallback
		}

		// Return format matching Ashcon's widely used format where possible, or just what our frontend needs
		return c.JSON(fiber.Map{
			"uuid":     uuid,
			"username": mojangResp.Name,
		})
	})

	app.Post("/api/instances/:id/type", func(c *fiber.Ctx) error {
		id := c.Params("id")
		inst, err := instanceManager.GetInstance(id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		var payload struct {
			Type    string `json:"type"`
			Version string `json:"version"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		// 1. Reset Instance
		if err := inst.Reset(payload.Type, payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to reset instance: %v", err)})
		}

		// 2. Install New Type (Reuse logic from install endpoint would be ideal, but for now we duplicate/call common logic)
		// We can call the same logic as the install endpoint.
		// Since we are in main, we can reuse the logic if we extract it, but for speed, I'll inline the essential install calls again.
		// NOTE: This duplicates logic from /api/instances/:id/install. Refactoring recommended later.

		if payload.Type == "custom" {
			// Nothing to install
			return c.JSON(fiber.Map{"status": "changed", "Note": "Custom type requires manual jar upload"})
		}

		vm := server.NewVersionsManager(inst.Manager)
		var installErr error
		var jarName string

		switch payload.Type {
		case "fabric":
			installErr = vm.InstallFabric(payload.Version)
			jarName = "fabric.jar"
		case "quilt":
			installErr = vm.InstallQuilt(payload.Version)
			jarName = "quilt.jar"
		case "forge":
			installErr = vm.InstallForge(payload.Version)
			jarName = "forge.jar"
		case "neoforge":
			installErr = vm.InstallNeoForge(payload.Version)
			jarName = "neoforge.jar"
		case "spigot":
			installErr = vm.InstallSpigot(payload.Version)
			jarName = "server.jar"
		default:
			return c.Status(400).JSON(fiber.Map{"error": "Unsupported type for auto-install"})
		}

		if installErr != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Reset successful, but install failed: %v", installErr)})
		}

		inst.SetJar(jarName)
		return c.JSON(fiber.Map{"status": "changed"})
	})

	// Instance Control Routes
	app.Post("/api/instances/:id/start", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		if err := inst.Start(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "started"})
	})

	app.Post("/api/instances/:id/stop", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		if err := inst.Stop(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "stopped"})
	})

	app.Post("/api/instances/:id/restart", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		if err := inst.Restart(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "restarting"})
	})

	app.Post("/api/instances/:id/command", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		var payload struct {
			Command string `json:"command"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}
		if err := inst.WriteCommand(payload.Command); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "sent"})
	})

	app.Delete("/api/instances/:id/mods", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		var payload struct {
			ProjectID string `json:"project_id"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		if err := inst.UninstallMod(payload.ProjectID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "uninstalled"})
	})

	app.Post("/api/instances/:id/install", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		var payload struct {
			Version string `json:"version"`
			Type    string `json:"type"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		// Create temporary version manager for this instance
		// We could refactor VersionsManager to take an *Instance or *Manager
		// Assuming VersionsManager currently takes *Manager
		vm := server.NewVersionsManager(inst.Manager)

		if payload.Type == "fabric" {
			// ... existing fabric logic ...
			// We need to support workDir in InstallFabric or context
			// InstallFabric in version.go likely assumes CWD or needs update.
			// Let's check server/versions.go content first.
			// For now, assuming we will fix it.
			if err := vm.InstallFabric(payload.Version); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			inst.SetJar("fabric.jar")
			return c.JSON(fiber.Map{"status": "installed", "jar": "fabric.jar"})
		}

		if payload.Type == "quilt" {
			if err := vm.InstallQuilt(payload.Version); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			inst.SetJar("quilt.jar")
			return c.JSON(fiber.Map{"status": "installed", "jar": "quilt.jar"})
		}

		if payload.Type == "forge" {
			if err := vm.InstallForge(payload.Version); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			inst.SetJar("forge.jar")
			return c.JSON(fiber.Map{"status": "installed", "jar": "forge.jar"})
		}

		if payload.Type == "neoforge" {
			if err := vm.InstallNeoForge(payload.Version); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			inst.SetJar("neoforge.jar")
			return c.JSON(fiber.Map{"status": "installed", "jar": "neoforge.jar"})
		}

		if payload.Type == "spigot" {
			if err := vm.InstallSpigot(payload.Version); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			inst.SetJar("server.jar")
			return c.JSON(fiber.Map{"status": "installed", "jar": "server.jar"})
		}

		return c.Status(400).JSON(fiber.Map{"error": "Unsupported version type"})
	})

	// File Management Routes
	app.Get("/api/instances/:id/files", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		path := c.Query("path", ".") // Default to root
		files, err := inst.ListFiles(path)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(files)
	})

	app.Get("/api/instances/:id/files/content", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		path := c.Query("path")
		if path == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Path required"})
		}
		content, err := inst.ReadFile(path)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		// Return text content
		return c.SendString(string(content))
	})

	app.Put("/api/instances/:id/files/content", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		var payload struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		if err := inst.WriteFile(payload.Path, []byte(payload.Content)); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "saved"})
	})

	app.Post("/api/instances/:id/files/upload", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		path := c.Query("path", ".")
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid form"})
		}

		files := form.File["files"]
		for _, file := range files {
			if err := inst.HandleUpload(path, file); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to upload %s: %v", file.Filename, err)})
			}
		}

		return c.JSON(fiber.Map{"status": "uploaded", "count": len(files)})
	})

	app.Delete("/api/instances/:id/files", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		path := c.Query("path")
		if path == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Path required"})
		}
		if err := inst.DeleteFile(path); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "deleted"})
	})

	app.Post("/api/instances/:id/files/mkdir", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}
		var payload struct {
			Path string `json:"path"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}
		if err := inst.Mkdir(payload.Path); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "created"})
	})

	// Mods Management Cache
	// TODO: Implement cache for search queries?

	app.Get("/api/instances/:id/mods/search", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		query := c.Query("query")
		typeFilter := c.Query("type", "mod") // "mod" or "modpack"
		isModpack := typeFilter == "modpack"
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		sort := c.Query("sort", "")
		sides := c.Query("sides", "")
		sidesList := []string{}
		if sides != "" {
			sidesList = strings.Split(sides, ",")
		}

		results, err := inst.SearchMods(query, isModpack, offset, sort, sidesList)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(results)
	})

	app.Get("/api/instances/:id/mods", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		ids, err := inst.GetInstalledMods()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ids)
	})

	app.Post("/api/instances/:id/mods", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		var payload struct {
			ProjectID string `json:"projectId"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		if err := inst.InstallMod(payload.ProjectID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "installed"})
	})

	app.Post("/api/instances/:id/modpacks", func(c *fiber.Ctx) error {
		inst, err := instanceManager.GetInstance(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
		}

		var payload struct {
			ProjectID string `json:"projectId"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		// Run in background as it might take time?
		// But for now, blocking to report error is safer for MVP, or use websocket for progress.
		// InstallModpack broadcasts progress, so we can return immediately or wait.
		// Let's return immediate Success and let WS handle it?
		// Or better, blocking so the UI loader spins until at least the process starts clearly.
		// Since InstallModpack does heavy IO, let's run it async and client watches console.

		go func() {
			if err := inst.InstallModpack(payload.ProjectID); err != nil {
				inst.Manager.Broadcast(fmt.Sprintf("Error installing modpack: %v", err))
			}
		}()

		return c.JSON(fiber.Map{"status": "installing"})
	})

	// WebSocket for Console
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/instances/:id/console", websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")
		inst, err := instanceManager.GetInstance(id)
		if err != nil {
			c.Close()
			return
		}

		inst.RegisterClient(c)
		defer inst.UnregisterClient(c)

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	// Serve Static Files (Frontend)
	app.Static("/", "./build")

	// SPA Fallback: Serve index.html for unknown routes
	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("./build/index.html")
	})

	// Network IP logging
	if ifaces, err := net.Interfaces(); err == nil {
		fmt.Println("Server available at:")
		fmt.Println("  http://localhost:3000")

		for _, i := range ifaces {
			// Skip down or loopback interfaces
			if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
				continue
			}

			// innovative heuristic filters
			name := strings.ToLower(i.Name)
			if strings.Contains(name, "docker") ||
				strings.Contains(name, "veth") ||
				strings.Contains(name, "br-") ||
				strings.Contains(name, "virbr") ||
				strings.Contains(name, "vmnet") ||
				strings.Contains(name, "wsl") {
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

				// Only support IPv4 for simplicity as requested
				if ip == nil || ip.To4() == nil {
					continue
				}

				fmt.Printf("  http://%s:3000\n", ip.String())
			}
		}
	}

	log.Fatal(app.Listen("0.0.0.0:3000"))
}
