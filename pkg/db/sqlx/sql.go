package sqlx

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func NewSQLDB(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	var dialector gorm.Dialector
	switch strings.ToUpper(config.Driver) {
	case DriverPostgres:
		dialector = postgres.Open(dsn)
	case DriverMySQL:
		dialector = mysql.Open(dsn)
	case DriverSQLite:
		dialector = sqlite.Open(dsn)
	case DriverSQLServer:
		dialector = sqlserver.Open(dsn)
	default:
		return nil, fmt.Errorf("[sqlx] unsupported driver: %s", config.Driver)
	}

	db, err := gorm.Open(dialector, &config.Config)
	if err != nil {
		return nil, fmt.Errorf("[sqlx] failed to open postgres connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("[sqlx] failed to get sql db: %w", err)
	}

	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}

	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 100
	}

	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 1 * time.Hour
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	log.Printf("[sqlx] connected to postgres: host=%s, port=%s, user=%s, db=%s\n", config.Host, config.Port, config.User, config.DBName)

	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("[sqlx] failed to get sql db: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("[sqlx] failed to close sql db: %w", err)
	}

	return nil
}
