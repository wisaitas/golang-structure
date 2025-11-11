package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/auth"
)

type AuthRouter struct {
	apiRouter   fiber.Router
	authUseCase *auth.UseCase
}

func NewAuthRouter(
	apiRouter fiber.Router,
	authUseCase *auth.UseCase,
) *AuthRouter {
	return &AuthRouter{
		apiRouter:   apiRouter,
		authUseCase: authUseCase,
	}
}

func (r *AuthRouter) Setup() {
	authRouter := r.apiRouter.Group("/auth")

	authRouter.Post("/register", r.authUseCase.Register.Handler)
}
