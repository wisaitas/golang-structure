package httpx

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

func RequestContext(c fiber.Ctx) context.Context {
	ctx, ok := c.Locals("requestContext").(context.Context)
	if ok && ctx != nil {
		return ctx
	}
	return c.Context()
}
