package getusers

import "github.com/gofiber/fiber/v2"

type Handler struct {
	service Service
}

func newHandler(
	service Service,
) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Handler(c *fiber.Ctx) error {
	return h.service.Service(c)
}
