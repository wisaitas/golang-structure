package initial

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/db/sqlx"
	"gorm.io/gorm"
)

type config struct {
	sqlDB *gorm.DB
}

func newConfig() *config {
	sqlDB, err := sqlx.NewSQLDB(golangstructure.Config.SQLDB)
	if err != nil {
		panic(err)
	}

	return &config{
		sqlDB: sqlDB,
	}
}
