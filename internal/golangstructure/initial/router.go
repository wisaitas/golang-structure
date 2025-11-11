package initial

import (
	"github.com/gofiber/fiber/v2"
	appRouter "github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/router"
)

type router struct {
	authRouter *appRouter.AuthRouter
	userRouter *appRouter.UserRouter
}

func newRouter(
	app *fiber.App,
	useCase *useCase,
) {
	apiRouter := app.Group("/api/v1")

	router := &router{
		authRouter: appRouter.NewAuthRouter(apiRouter, useCase.authUseCase),
		userRouter: appRouter.NewUserRouter(apiRouter, useCase.userUseCase),
	}

	router.setup()
}

func (r *router) setup() {
	r.authRouter.Setup()
	r.userRouter.Setup()
}
