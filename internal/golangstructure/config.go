package golangstructure

import "github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/db/sqlx"

var Config struct {
	Service struct {
		Name        string `env:"SERVICE_NAME" envDefault:"golang-structure"`
		Port        int    `env:"SERVICE_PORT" envDefault:"8080"`
		Stage       string `env:"SERVICE_STAGE" envDefault:"dev"`
		MaskPattern string `env:"MASK_PATTERN" envDefault:"{}"`
	}
	SQLDB sqlx.Config

	Bcrypt struct {
		Cost int `env:"BCRYPT_COST" envDefault:"10"`
	}
}
