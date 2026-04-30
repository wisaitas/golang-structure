package register_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	repositoryMocks "github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository/mocks"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/auth/register"
	bcryptxMock "github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/bcryptx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/db/gormx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
	"gorm.io/gorm"
)

type ServiceTestSuite struct {
	suite.Suite
	ctx         context.Context
	request     *register.Request
	userRepo    *repositoryMocks.MockUserRepository
	userLogRepo *repositoryMocks.MockUserLogRepository
	bcrypt      *bcryptxMock.MockBcrypt
	service     register.Service
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	golangstructure.Config.Bcrypt.Cost = 4
	s.ctx = context.Background()
	s.request = &register.Request{
		Name:            "john",
		Email:           "john@doe.com",
		Age:             20,
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	s.userRepo = repositoryMocks.NewMockUserRepository(s.T())
	s.userLogRepo = repositoryMocks.NewMockUserLogRepository(s.T())
	s.bcrypt = bcryptxMock.NewMockBcrypt(s.T())
	s.service = register.NewService(s.userRepo, s.userLogRepo, s.bcrypt)
}

func (s *ServiceTestSuite) TestServiceSuccess() {
	txDB := &gorm.DB{}
	transactionFnCalled := false
	s.bcrypt.EXPECT().GenerateFromPassword(s.request.Password, golangstructure.Config.Bcrypt.Cost).
		Return([]byte("hashed-password"), nil).Once()

	s.userRepo.EXPECT().
		Transaction(s.ctx, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(gormx.BaseRepository[entity.User]) error) error {
			transactionFnCalled = true
			return fn(s.userRepo)
		})

	s.userRepo.EXPECT().
		Create(s.ctx, mock.AnythingOfType("*entity.User")).
		RunAndReturn(func(ctx context.Context, value interface{}) error {
			user := value.(*entity.User)
			user.ID = 10
			return nil
		})

	s.userRepo.EXPECT().
		GetDB(s.ctx).
		Return(txDB)

	s.userLogRepo.EXPECT().
		WithTx(txDB).
		Return(s.userLogRepo)

	s.userLogRepo.EXPECT().
		Create(s.ctx, mock.AnythingOfType("*entity.UserLog")).
		RunAndReturn(func(ctx context.Context, value interface{}) error {
			userLog := value.(*entity.UserLog)
			s.Equal(10, userLog.UserID)
			s.Equal("register", userLog.Action)
			return nil
		})

	err := s.service.Service(s.ctx, s.request)
	s.NoError(err)
	s.True(transactionFnCalled)
}

func (s *ServiceTestSuite) TestServiceCreateUserDuplicateKey() {
	s.bcrypt.EXPECT().GenerateFromPassword(s.request.Password, golangstructure.Config.Bcrypt.Cost).
		Return([]byte("hashed-password"), nil).Once()

	s.userRepo.EXPECT().
		Transaction(s.ctx, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(gormx.BaseRepository[entity.User]) error) error {
			return fn(s.userRepo)
		})

	s.userRepo.EXPECT().
		Create(s.ctx, mock.AnythingOfType("*entity.User")).
		Return(errors.New("duplicate key value violates unique constraint"))

	err := s.service.Service(s.ctx, s.request)
	s.Error(err)
	s.Equal(http.StatusConflict, httpx.StatusCodeFromError(err, http.StatusOK))
	s.Equal(httpx.CodeConflict, httpx.ResponseCodeFromError(err))
}

func (s *ServiceTestSuite) TestServiceCreateUserLogError() {
	txDB := &gorm.DB{}
	s.bcrypt.EXPECT().GenerateFromPassword(s.request.Password, golangstructure.Config.Bcrypt.Cost).
		Return([]byte("hashed-password"), nil).Once()

	s.userRepo.EXPECT().
		Transaction(s.ctx, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(gormx.BaseRepository[entity.User]) error) error {
			return fn(s.userRepo)
		})

	s.userRepo.EXPECT().
		Create(s.ctx, mock.AnythingOfType("*entity.User")).
		RunAndReturn(func(ctx context.Context, value interface{}) error {
			user := value.(*entity.User)
			user.ID = 7
			return nil
		})

	s.userRepo.EXPECT().
		GetDB(s.ctx).
		Return(txDB)

	s.userLogRepo.EXPECT().
		WithTx(txDB).
		Return(s.userLogRepo)

	s.userLogRepo.EXPECT().
		Create(s.ctx, mock.AnythingOfType("*entity.UserLog")).
		Return(errors.New("insert user log failed"))

	err := s.service.Service(s.ctx, s.request)
	s.Error(err)
	s.Equal(http.StatusInternalServerError, httpx.StatusCodeFromError(err, http.StatusOK))
	s.Equal(httpx.CodeInternal, httpx.ResponseCodeFromError(err))
}

func (s *ServiceTestSuite) TestServiceTransactionError() {
	s.bcrypt.EXPECT().GenerateFromPassword(s.request.Password, golangstructure.Config.Bcrypt.Cost).
		Return([]byte("hashed-password"), nil).Once()

	s.userRepo.EXPECT().
		Transaction(s.ctx, mock.Anything).
		Return(errors.New("begin tx failed")).Once()

	err := s.service.Service(s.ctx, s.request)
	s.Error(err)
	s.Equal(http.StatusInternalServerError, httpx.StatusCodeFromError(err, http.StatusOK))
	s.Equal(httpx.CodeInternal, httpx.ResponseCodeFromError(err))
}

func (s *ServiceTestSuite) TestServiceBcryptError() {
	s.bcrypt.EXPECT().GenerateFromPassword(s.request.Password, golangstructure.Config.Bcrypt.Cost).
		Return(nil, errors.New("bcrypt failed")).Once()

	err := s.service.Service(s.ctx, s.request)
	s.Error(err)
	s.Equal(http.StatusInternalServerError, httpx.StatusCodeFromError(err, http.StatusOK))
	s.Equal(httpx.CodeInternal, httpx.ResponseCodeFromError(err))
}
