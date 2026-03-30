package initial

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/middleware"
)

func newMiddleware(app *fiber.App) {
	app.Use(middleware.Prometheus(app))
	app.Use(middleware.Logger())
	app.Use(middleware.Cors())
}
