package web

import (
	"jjmc/auth"
	"jjmc/instances"
	"jjmc/web/handlers"
	"jjmc/web/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func RegisterRoutes(app *fiber.App, authManager *auth.AuthManager, instanceManager *instances.InstanceManager) {
	// Middleware
	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(middleware.AuthMiddleware(authManager))

	// Handlers
	authHandler := handlers.NewAuthHandler(authManager)
	systemHandler := handlers.NewSystemHandler()
	instHandler := handlers.NewInstanceHandler(instanceManager)

	// Auth Routes
	authGroup := app.Group("/api/auth")
	authGroup.Get("/status", authHandler.GetStatus)
	authGroup.Post("/setup", authHandler.Setup)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/logout", authHandler.Logout)

	// System Routes
	sysGroup := app.Group("/api/system")
	sysGroup.Get("/files", systemHandler.GetFiles)
	sysGroup.Get("/uuid", systemHandler.GetUUID)

	// Modrinth Helper Routes (System-wide)
	verGroup := app.Group("/api/versions")
	verGroup.Get("/game", systemHandler.GetGameVersions)
	verGroup.Get("/loader", systemHandler.GetLoaders)

	// Instance Routes
	instGroup := app.Group("/api/instances")
	instGroup.Get("/", instHandler.List)
	instGroup.Post("/", instHandler.Create)
	instGroup.Post("/import", instHandler.Import)

	// Single Instance Routes
	inst := instGroup.Group("/:id")
	inst.Get("/", instHandler.Get)
	inst.Delete("/", instHandler.Delete)
	inst.Patch("/", instHandler.UpdateSettings)
	inst.Post("/type", instHandler.ChangeType)
	inst.Post("/start", instHandler.Start)
	inst.Post("/stop", instHandler.Stop)
	inst.Post("/restart", instHandler.Restart)
	inst.Post("/command", instHandler.Command)
	inst.Post("/install", instHandler.Install)

	// Instance Files
	files := inst.Group("/files")
	files.Get("/", instHandler.ListFiles)
	files.Get("/content", instHandler.ReadFile)
	files.Put("/content", instHandler.WriteFile)
	files.Post("/upload", instHandler.Upload)
	files.Delete("/", instHandler.DeleteFile)
	files.Post("/mkdir", instHandler.Mkdir)
	files.Post("/compress", instHandler.Compress)
	files.Post("/decompress", instHandler.Decompress)

	// Instance Mods
	mods := inst.Group("/mods")
	mods.Get("/", instHandler.GetInstalledMods)
	mods.Post("/", instHandler.InstallMod)
	mods.Delete("/", instHandler.UninstallMod)
	mods.Get("/search", instHandler.SearchMods)

	inst.Post("/modpacks", instHandler.InstallModpack)

	// WebSocket
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

		inst.Manager.RegisterClient(c)
		defer inst.Manager.UnregisterClient(c)

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	// Backup Routes
	RegisterBackupRoutes(app, authManager, instanceManager)
}
