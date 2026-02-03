package web

import (
	"jjmc/auth"
	"jjmc/instances"
	"jjmc/services/java_manager"
	"jjmc/services/scheduler"
	"jjmc/web/handlers"
	"jjmc/web/middleware"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RegisterRoutes(app *fiber.App, authManager *auth.AuthManager, instanceManager *instances.InstanceManager, scheduler *scheduler.Scheduler, javaManager *java_manager.JavaManager) {

	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(helmet.New(helmet.Config{
		ContentSecurityPolicy:     "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data: https://cdn.modrinth.com https://git.io https://avatars.githubusercontent.com https://static.spigotmc.org https://www.spigotmc.org https://secure.gravatar.com https://minotar.net https://i.imgur.com; font-src 'self' data:; connect-src 'self' ws: wss:;",
		CrossOriginEmbedderPolicy: "unsafe-none",
	}))
	app.Use(middleware.AuthMiddleware(authManager))

	// Rate Limiter for Login
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many login attempts, please try again later",
			})
		},
	})

	authHandler := handlers.NewAuthHandler(authManager)
	systemHandler := handlers.NewSystemHandler()
	// Folder Routes
	folderHandler := &handlers.FolderHandler{}
	app.Get("/api/folders", folderHandler.List)
	app.Post("/api/folders", folderHandler.Create)
	app.Delete("/api/folders/:id", folderHandler.Delete)
	app.Patch("/api/folders/:id", folderHandler.Rename)

	instHandler := handlers.NewInstanceHandler(instanceManager)
	scheduleHandler := handlers.NewScheduleHandler(scheduler)

	authGroup := app.Group("/api/auth")
	authGroup.Get("/status", authHandler.GetStatus)
	authGroup.Post("/setup", authHandler.Setup)
	authGroup.Post("/login", loginLimiter, authHandler.Login)
	authGroup.Post("/logout", authHandler.Logout)

	sysGroup := app.Group("/api/system")
	sysGroup.Get("/files", systemHandler.GetFiles)
	sysGroup.Get("/uuid", systemHandler.GetUUID)

	verGroup := app.Group("/api/versions")
	verGroup.Get("/game", systemHandler.GetGameVersions)
	verGroup.Get("/loader", systemHandler.GetLoaders)

	modpackHandler := handlers.NewModpackHandler()
	mpGroup := app.Group("/api/modpacks")
	mpGroup.Get("/search", modpackHandler.Search)

	javaHandler := handlers.NewJavaHandler(javaManager)
	javaGroup := app.Group("/api/java")
	javaGroup.Get("/installed", javaHandler.ListInstalled)
	javaGroup.Post("/install", javaHandler.Install)
	javaGroup.Delete("/:name", javaHandler.Delete)

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

	// Schedules
	schedules := inst.Group("/schedules")
	schedules.Get("/", scheduleHandler.List)
	schedules.Post("/", scheduleHandler.Create)
	schedules.Put("/:scheduleId", scheduleHandler.Update)
	schedules.Delete("/:scheduleId", scheduleHandler.Delete)

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

	app.Get("/ws/instances/:id/stats", websocket.New(instHandler.StatsWebSocket))

	RegisterBackupRoutes(app, authManager, instanceManager)
	handlers.RegisterNetworkRoutes(app, instanceManager)
}
