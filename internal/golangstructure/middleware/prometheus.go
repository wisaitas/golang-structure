package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/promx"
)

func Prometheus(app *fiber.App) fiber.Handler {
	return promx.NewMiddleware(app, promx.Config{
		ServiceName: golangstructure.Config.Service.Name,
	})
}
