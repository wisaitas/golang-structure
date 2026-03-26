package bcryptx

import (
	"fmt"

	bcryptLib "golang.org/x/crypto/bcrypt"
)

type Bcrypt interface {
	GenerateFromPassword(password string, cost int) ([]byte, error)
	CompareHashAndPassword(hashedPassword, password []byte) error
}

type bcrypt struct {
}

func NewBcrypt() Bcrypt {
	return &bcrypt{}
}

func (r *bcrypt) GenerateFromPassword(password string, cost int) ([]byte, error) {
	hashedPassword, err := bcryptLib.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return nil, fmt.Errorf("[bcrypt] : %w", err)
	}

	return hashedPassword, nil
}

func (r *bcrypt) CompareHashAndPassword(hashedPassword, password []byte) error {
	if err := bcryptLib.CompareHashAndPassword(hashedPassword, password); err != nil {
		return fmt.Errorf("[bcrypt] : %w", err)
	}

	return nil
}
