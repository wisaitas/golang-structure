package createuser

import "github.com/gofiber/fiber/v2"

type Handler interface {
	CreateUser(c *fiber.Ctx) error
}

type handler struct {
	service service
}

func newHandler(
	service service,
) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateUser(c *fiber.Ctx) error {
	req := Request{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return h.service.CreateUser(c, &req)
}
