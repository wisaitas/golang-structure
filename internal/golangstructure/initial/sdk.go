package initial

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/bcryptx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/validatorx"
)

type sdk struct {
	validator validatorx.Validator
	bcrypt    bcryptx.Bcrypt
}

func newSDK() *sdk {
	return &sdk{
		validator: validatorx.NewValidator(),
		bcrypt:    bcryptx.NewBcrypt(),
	}
}
