package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
)

func Logger() fiber.Handler {
	return httpx.NewLogger(httpx.LoggerConfig{
		ServiceName:    golangstructure.Config.Service.Name,
		MaskMapPattern: golangstructure.Config.Service.MaskPattern,
	})
}
