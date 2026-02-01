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

	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(middleware.AuthMiddleware(authManager))

	authHandler := handlers.NewAuthHandler(authManager)
	systemHandler := handlers.NewSystemHandler()
	instHandler := handlers.NewInstanceHandler(instanceManager)

	authGroup := app.Group("/api/auth")
	authGroup.Get("/status", authHandler.GetStatus)
	authGroup.Post("/setup", authHandler.Setup)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/logout", authHandler.Logout)

	sysGroup := app.Group("/api/system")
	sysGroup.Get("/files", systemHandler.GetFiles)
	sysGroup.Get("/uuid", systemHandler.GetUUID)

	verGroup := app.Group("/api/versions")
	verGroup.Get("/game", systemHandler.GetGameVersions)
	verGroup.Get("/loader", systemHandler.GetLoaders)

	instGroup := app.Group("/api/instances")
	instGroup.Get("/", instHandler.List)
	instGroup.Post("/", instHandler.Create)
	instGroup.Post("/import", instHandler.Import)

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

	files := inst.Group("/files")
	files.Get("/", instHandler.ListFiles)
	files.Get("/content", instHandler.ReadFile)
	files.Put("/content", instHandler.WriteFile)
	files.Post("/upload", instHandler.Upload)
	files.Delete("/", instHandler.DeleteFile)
	files.Post("/mkdir", instHandler.Mkdir)
	files.Post("/compress", instHandler.Compress)
	files.Post("/decompress", instHandler.Decompress)

	mods := inst.Group("/mods")
	mods.Get("/", instHandler.GetInstalledMods)
	mods.Post("/", instHandler.InstallMod)
	mods.Delete("/", instHandler.UninstallMod)
	mods.Get("/search", instHandler.SearchMods)
	mods.Get("/:projectId/versions", instHandler.GetModVersions)

	tunnel := inst.Group("/tunnel")
	tunnel.Get("/", instHandler.GetTunnelStatus)
	tunnel.Post("/start", instHandler.StartTunnel)
	tunnel.Post("/stop", instHandler.StopTunnel)

	inst.Post("/modpacks", instHandler.InstallModpack)

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

	RegisterBackupRoutes(app, authManager, instanceManager)
}
