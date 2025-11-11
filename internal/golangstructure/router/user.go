package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user"
)

type UserRouter struct {
	apiRouter   fiber.Router
	userUseCase *user.UseCase
}

func NewUserRouter(
	apiRouter fiber.Router,
	userUseCase *user.UseCase,
) *UserRouter {
	return &UserRouter{
		apiRouter:   apiRouter,
		userUseCase: userUseCase,
	}
}

func (r *UserRouter) Setup() {
	userRouter := r.apiRouter.Group("/users")

	userRouter.Get("/", r.userUseCase.GetUsers.Handler)

	userRouter.Post("/", r.userUseCase.CreateUser.Handler)

	userRouter.Put("/:user_id", r.userUseCase.UpdateUser.Handler)

	userRouter.Delete("/:user_id", r.userUseCase.DeleteUser.Handler)
}
