package getusers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/golang-structure/pkg/httpx"
	"github.com/wisaitas/golang-structure/pkg/validatorx"
)

type Handler struct {
	operation string
	service   Service
	validator validatorx.Validator
}

func newHandler(
	service Service,
	validator validatorx.Validator,
) *Handler {
	return &Handler{
		operation: "[getusers.handler]",
		service:   service,
		validator: validator,
	}
}

func (h *Handler) Handler(c fiber.Ctx) error {
	req := new(Request)
	if err := h.validator.ValidateStruct(req); err != nil {
		return httpx.NewErrorResponse[any](
			c,
			fiber.StatusBadRequest,
			httpx.CodeBadRequest,
			httpx.WrapErrorWithCode(h.operation, err, fiber.StatusBadRequest, httpx.CodeBadRequest),
			nil,
			"",
		)
	}

	data, err := h.service.Service(httpx.RequestContext(c), req)
	if err != nil {
		return httpx.NewErrorResponse[any](c, 0, "", err, nil, h.operation)
	}

	return httpx.NewSuccessResponse[[]Response](c, &data, fiber.StatusOK, httpx.CodeOK, nil, nil)
}
