package bootstrap

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/opentreehole/go-common"

	"treehole_next/apis"
	"treehole_next/apis/hole"
	"treehole_next/apis/message"
	"treehole_next/config"
	"treehole_next/models"
	"treehole_next/utils"
	"treehole_next/utils/sensitive"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Init() (*fiber.App, context.CancelFunc) {
	config.InitConfig()
	utils.InitCache()
	sensitive.InitSensitiveLabelMap()
	models.Init()
	models.InitDB()
	models.InitAdminList()

	app := fiber.New(fiber.Config{
		ErrorHandler:          common.ErrorHandler,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	})
	registerMiddlewares(app)
	apis.RegisterRoutes(app)

	return app, startTasks()
}

func registerMiddlewares(app *fiber.App) {
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(common.MiddlewareGetUserID)
	if config.Config.Mode != "bench" {
		app.Use(common.MiddlewareCustomLogger)
	}
	app.Use(pprof.New())
}

func startTasks() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go hole.UpdateHoleViews(ctx)
	go hole.PurgeHole(ctx)
	go message.PurgeMessage()
	// go models.UpdateAdminList(ctx)
	go sensitive.UpdateSensitiveLabelMap(ctx)
	return cancel
}
