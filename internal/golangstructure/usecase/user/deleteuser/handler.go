package deleteuser

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

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
	param := c.Params("user_id")
	userID, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return h.service.Service(c, userID)
}
