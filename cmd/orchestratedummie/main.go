package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"

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
	backendURL := getEnv("BACKEND_URL", "http://localhost:8080")
	app := fiber.New()

	app.Use(httpx.NewLogger(httpx.LoggerConfig{
		ServiceName:    "orchestrate-service",
		MaskMapPattern: "{}",
	}))

	app.Post("/register", func(c fiber.Ctx) error {
		req := make(map[string]any)
		if err := c.Bind().Body(&req); err != nil {
			return httpx.NewErrorResponse[any](
				c,
				fiber.StatusBadRequest,
				httpx.CodeBadRequest,
				httpx.WrapErrorWithCode("[main.post]", err, fiber.StatusBadRequest, httpx.CodeBadRequest),
				nil,
				"[main.post]",
			)
		}

		resp := new(httpx.StandardResponse[any])
		if err := httpx.Client(
			c,
			http.MethodPost,
			backendURL+"/api/v1/auth/register",
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
				httpx.WrapError("[orchestrate]", err, statusCode),
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
				httpx.WrapError("[orchestrate]", errors.New("downstream returned status "+strconv.Itoa(resp.StatusCode)), resp.StatusCode),
				nil,
				"",
			)
		}

		return c.Status(resp.StatusCode).JSON(resp)
	})

	app.Listen(":8081")
}
