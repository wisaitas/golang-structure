package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	orchestrateURL := getEnv("ORCHESTRATE_URL", "http://localhost:8081")
	app := fiber.New()

	app.Use(httpx.NewLogger(httpx.LoggerConfig{
		ServiceName:    "gateway-service",
		MaskMapPattern: "{}",
	}))

	app.Post("/register", func(c fiber.Ctx) error {
		req := make(map[string]any)
		if err := c.Bind().Body(&req); err != nil {
			return httpx.NewErrorResponse[any](
				c,
				fiber.StatusBadRequest,
				httpx.CodeBadRequest,
				httpx.WrapError("dummy.register.bind_body", err, fiber.StatusBadRequest),
				nil,
				"",
			)
		}

		resp := new(httpx.StandardResponse[any])
		if err := httpx.Client(
			c,
			http.MethodPost,
			orchestrateURL+"/register",
			req,
			resp,
		); err != nil {
			statusCode := resp.StatusCode
			if statusCode == 0 {
				statusCode = fiber.StatusBadGateway
			}
			return httpx.NewErrorResponse[any](
				c,
				statusCode,
				httpx.CodeBadGateway,
				httpx.WrapError("dummy.register.call_golang_structure", err, statusCode),
				nil,
				"",
			)
		}
		if !httpx.CheckStatusCode2xx(resp.StatusCode) {
			apiCode := resp.Code
			if apiCode == "" {
				apiCode = httpx.CodeForHTTPStatus(resp.StatusCode)
			}
			return httpx.NewErrorResponse[any](
				c,
				resp.StatusCode,
				apiCode,
				httpx.WrapError("dummy.register.call_golang_structure", fmt.Errorf("downstream returned status %d", resp.StatusCode), resp.StatusCode),
				nil,
				"",
			)
		}

		return c.Status(resp.StatusCode).JSON(resp)
	})

	app.Listen(":3000")
}
