package initial

import (
	"github.com/gofiber/fiber/v2"
	appRouter "github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/router"
)

type router struct {
	userRouter *appRouter.UserRouter
}

func newRouter(
	app *fiber.App,
	strategy *strategy,
) {
	apiRouter := app.Group("/api/v1")

	router := &router{
		userRouter: appRouter.NewUserRouter(apiRouter, strategy.userStrategy),
	}

	router.setup()
}

func (r *router) setup() {
	r.userRouter.Setup()
}
