package initial

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/postgres"

	"gorm.io/gorm"
)

type config struct {
	postgresDB *gorm.DB
}

func newConfig() *config {
	postgresDB := postgres.ConnectDB(postgres.Config{
		Host:     golangstructure.ENV.Postgres.Host,
		Port:     golangstructure.ENV.Postgres.Port,
		User:     golangstructure.ENV.Postgres.User,
		Password: golangstructure.ENV.Postgres.Password,
		DBName:   golangstructure.ENV.Postgres.DBName,
		SSLMode:  golangstructure.ENV.Postgres.SSLMode,
	})

	if err := postgresDB.AutoMigrate(&entity.User{}); err != nil {
		panic(err)
	}

	return &config{
		postgresDB: postgresDB,
	}
}
