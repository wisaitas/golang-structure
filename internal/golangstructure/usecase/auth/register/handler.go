package register

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
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
	req := new(Request)
	if err := c.Bind().Body(req); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err, nil)
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err, nil)
	}

	resp := h.service.Service(c, req)

	return c.JSON(resp)
}
