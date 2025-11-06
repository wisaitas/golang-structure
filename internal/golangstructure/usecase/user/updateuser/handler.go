package updateuser

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	UpdateUser(c *fiber.Ctx) error
}

type handler struct {
	service   service
	validator *validator.Validate
}

func newHandler(
	service service,
	validator *validator.Validate,
) Handler {
	return &handler{
		service:   service,
		validator: validator,
	}
}

func (h *handler) UpdateUser(c *fiber.Ctx) error {
	req := Request{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	param := c.Params("user_id")
	userID, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return h.service.UpdateUser(c, &req, userID)
}
