package register

import "github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"

func (s *service) mapRequestToEntity(request *Request) *entity.User {
	return &entity.User{
		Name:     request.Name,
		Email:    request.Email,
		Age:      request.Age,
		Password: request.Password,
	}
}
