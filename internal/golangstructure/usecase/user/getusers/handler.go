package getusers

import "github.com/gofiber/fiber/v2"

type Handler interface {
	GetUsers(c *fiber.Ctx) error
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

func (h *handler) GetUsers(c *fiber.Ctx) error {
	return h.service.GetUsers(c)
}
