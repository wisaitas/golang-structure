package deleteuser

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	DeleteUser(c *fiber.Ctx) error
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

func (h *handler) DeleteUser(c *fiber.Ctx) error {
	param := c.Params("user_id")
	userID, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return h.service.DeleteUser(c, userID)
}
