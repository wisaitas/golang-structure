package getusers

import "github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"

func (s *service) mapEntitiesToResponses(users []entity.TblUsers) []Response {
	out := make([]Response, 0, len(users))
	for _, u := range users {
		out = append(out, Response{
			ID:   u.ID,
			Name: u.Name,
			Age:  u.Age,
		})
	}
	return out
}
