package updateuser

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/validatorx"
)

type Handler struct {
	service   Service
	validator validatorx.Validator
}

func newHandler(
	service Service,
	validator validatorx.Validator,
) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

func (h *Handler) Handler(c fiber.Ctx) error {
	req := Request{}
	if err := c.Bind().Body(&req); err != nil {
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

	if err := h.validator.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return h.service.Service(c, &req, userID)
}
