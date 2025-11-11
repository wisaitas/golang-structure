package initial

import "github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/validatorx"

type sdk struct {
	validator validatorx.Validator
}

func newSDK() *sdk {
	return &sdk{
		validator: validatorx.NewValidator(),
	}
}
