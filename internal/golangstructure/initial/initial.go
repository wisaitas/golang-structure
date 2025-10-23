package initial

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	FiberApp *fiber.App
	config   config
}

func New() *App {
	config := NewConfig()
	app := fiber.New()

	return &App{
		FiberApp: app,
		config:   config,
	}
}

func (a *App) Run() {
	go func() {
		if err := a.FiberApp.Listen(":" + golangstructure.ENV.Port); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func (a *App) Shutdown() {
	sqlDB, err := a.config.postgresDB.DB()
	if err != nil {
		panic(err)
	}

	if err := sqlDB.Close(); err != nil {
		panic(err)
	}

	fmt.Println("Shutting down...")
}
