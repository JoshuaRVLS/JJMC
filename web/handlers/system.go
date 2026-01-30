package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type SystemHandler struct{}

type SystemFile struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) GetFiles(c *fiber.Ctx) error {
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
}

func (h *SystemHandler) GetGameVersions(c *fiber.Ctx) error {
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
}

func (h *SystemHandler) GetLoaders(c *fiber.Ctx) error {
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
}

func (h *SystemHandler) GetUUID(c *fiber.Ctx) error {
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

	return c.JSON(fiber.Map{
		"uuid":     uuid,
		"username": mojangResp.Name,
	})
}
