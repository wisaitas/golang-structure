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
		return httpx.NewErrorResponse[any](
			c,
			fiber.StatusBadRequest,
			httpx.WrapError("register.handler.bind_body", err, fiber.StatusBadRequest),
			nil,
		)
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		return httpx.NewErrorResponse[any](
			c,
			fiber.StatusBadRequest,
			httpx.WrapError("register.handler.validate", err, fiber.StatusBadRequest),
			nil,
		)
	}

	if err := h.service.Service(httpx.RequestContext(c), req); err != nil {
		statusCode := httpx.StatusCodeFromError(err, fiber.StatusInternalServerError)
		err = httpx.WrapError("register.handler.service", err, statusCode)
		return httpx.NewErrorResponse[any](c, statusCode, err, nil)
	}

	return httpx.NewSuccessResponse[any](c, nil, fiber.StatusCreated, nil, nil)
}
