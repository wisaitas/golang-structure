package initial

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/db/sqlx"

	"github.com/gofiber/fiber/v3"
)

func init() {
	for _, path := range []string{".env", "../.env", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}

	if err := env.Parse(&golangstructure.Config); err != nil {
		log.Fatalln(err)
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
	app := fiber.New(fiber.Config{
		AppName: golangstructure.Config.Service.Name,
	})
	newMiddleware(app, config)

	newRouter(app, useCase)

	return &App{
		FiberApp: app,
		config:   config,
	}
}

func (a *App) Run() {
	go func() {
		if err := a.FiberApp.Listen(fmt.Sprintf(":%d", golangstructure.Config.Service.Port)); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func (a *App) Shutdown() {
	if err := sqlx.Close(a.config.sqlDB); err != nil {
		panic(err)
	}

	fmt.Println("Shutting down...")
}
