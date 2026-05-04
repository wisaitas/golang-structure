package initial

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/golang-structure/pkg/bcryptx"
	"github.com/wisaitas/golang-structure/pkg/logx"
	"github.com/wisaitas/golang-structure/pkg/validatorx"
)

type sdk struct {
	validator validatorx.Validator
	bcrypt    bcryptx.Bcrypt
	logger    logx.Logger
}

func newSDK() *sdk {
	return &sdk{
		validator: validatorx.NewValidator(),
		bcrypt:    bcryptx.NewBcrypt(),
		logger:    logx.NewLogger(golangstructure.Config.Log.Level),
	}
}
