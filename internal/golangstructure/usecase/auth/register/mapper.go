package register

import "github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"

func (s *service) mapRequestToEntity(request *Request) *entity.TblUsers {
	return &entity.TblUsers{
		Name:     request.Name,
		Email:    request.Email,
		Age:      request.Age,
		Password: request.Password,
	}
}

func (s *service) mapRequestToUserLog(user *entity.TblUsers) *entity.TblUserLogs {
	return &entity.TblUserLogs{
		UserID: user.ID,
		Action: "register",
	}
}
