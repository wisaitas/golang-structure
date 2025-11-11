package initial

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v11"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"

	"github.com/gofiber/fiber/v2"
)

func init() {
	if err := env.Parse(&golangstructure.ENV); err != nil {
		panic(err)
	}
}

type App struct {
	FiberApp *fiber.App
	config   *config
}

func New() *App {
	config := newConfig()
	sdk := newSDK()
	repository := newRepository(config)
	useCase := newUseCase(repository, sdk)
	app := fiber.New()
	newMiddleware(app)

	newRouter(app, useCase)

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
