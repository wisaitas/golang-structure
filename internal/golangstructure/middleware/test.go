package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// logger
// CORS
// Rate limiter ???

func MiddlewareTest() fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("middleware test")
		return c.Next()
	}
}
