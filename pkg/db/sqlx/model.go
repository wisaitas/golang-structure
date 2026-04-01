package sqlx

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseEntity struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Version   int             `gorm:"type:integer;not null;default:0"`
	CreatedAt time.Time       `gorm:"type:timestamp;not null;default:now()"`
	CreatedBy *uuid.UUID      `gorm:"type:uuid"`
	UpdatedAt time.Time       `gorm:"type:timestamp;not null;default:now()"`
	UpdatedBy *uuid.UUID      `gorm:"type:uuid"`
	DeletedAt *gorm.DeletedAt `gorm:"type:timestamp"`
	DeletedBy *uuid.UUID      `gorm:"type:uuid"`
}

type Config struct {
	Host            string        `env:"SQLDB_HOST"`
	Port            string        `env:"SQLDB_PORT"`
	User            string        `env:"SQLDB_USER"`
	Password        string        `env:"SQLDB_PASSWORD"`
	DBName          string        `env:"SQLDB_DB_NAME"`
	SSLMode         string        `env:"SQLDB_SSL_MODE"`
	MaxIdleConns    int           `env:"SQLDB_MAX_IDLE_CONNS"`
	MaxOpenConns    int           `env:"SQLDB_MAX_OPEN_CONNS"`
	ConnMaxLifetime time.Duration `env:"SQLDB_CONN_MAX_LIFETIME"`
	Driver          string        `env:"SQLDB_DRIVER"`
	MaskPattern     string        `env:"MASK_PATTERN"`
	gorm.Config     `env:"-"`
}
