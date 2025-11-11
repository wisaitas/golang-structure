package createuser

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
	req := Request{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return h.service.Service(c, &req)
}
