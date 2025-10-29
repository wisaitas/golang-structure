package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user"
)

type UserRouter struct {
	apiRouter    fiber.Router
	userStrategy user.Strategy
}

func NewUserRouter(
	apiRouter fiber.Router,
	userStrategy user.Strategy,
) *UserRouter {
	return &UserRouter{
		apiRouter:    apiRouter,
		userStrategy: userStrategy,
	}
}

func (r *UserRouter) Setup() {
	userRouter := r.apiRouter.Group("/users")

	userRouter.Get("/", r.userStrategy.GetUsers)

	userRouter.Post("/", r.userStrategy.CreateUser)

	userRouter.Put("/:id", r.userStrategy.UpdateUser)
}
