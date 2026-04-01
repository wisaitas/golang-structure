package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/db/sqlx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
	"gorm.io/gorm"
)

type healthCheckResult struct {
	Checks map[string]string `json:"checks"`
}

func Healthz(app *fiber.App, sqlDB *gorm.DB) fiber.Handler {
	app.Get("/livez", func(c fiber.Ctx) error {
		return httpx.NewSuccessResponse[any](c, nil, fiber.StatusOK, httpx.CodeOK, nil, nil)
	})

	app.Get("/readyz", func(c fiber.Ctx) error {
		if err := sqlx.Ping(sqlDB); err != nil {
			return httpx.NewErrorResponse[any](
				c,
				fiber.StatusServiceUnavailable,
				httpx.CodeServiceUnavailable,
				fmt.Errorf("database ping failed: %w", err),
				nil,
				"",
			)
		}

		data := healthCheckResult{
			Checks: map[string]string{"database": "ok"},
		}
		return httpx.NewSuccessResponse(c, &data, fiber.StatusOK, httpx.CodeOK, nil, nil)
	})

	return func(c fiber.Ctx) error {
		return c.Next()
	}
}
