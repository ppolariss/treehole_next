package report

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App) {
	app.Get("/reports/:id", GetReport)
	app.Get("/reports", ListReports)
	app.Post("/reports", AddReport)
	app.Delete("/reports/:id", DeleteReport)
}