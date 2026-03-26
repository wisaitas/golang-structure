package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
)

func main() {
	app := fiber.New()

	app.Use(httpx.NewLogger("dummy-service"))

	app.Listen(":3000")
}
