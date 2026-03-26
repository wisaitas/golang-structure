package sqlx

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/mask"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	gcfg := config.Config
	if m := mask.ParsePatternMap(config.MaskPattern); len(m) > 0 {
		gcfg.Logger = logger.New(
			log.New(&sqlLogMaskWriter{w: os.Stdout, m: m}, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				Colorful:                  true,
				IgnoreRecordNotFoundError: true,
			},
		)
	}

	db, err := gorm.Open(dialector, &gcfg)
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

type sqlLogMaskWriter struct {
	w io.Writer
	m map[string]string
}

func (w *sqlLogMaskWriter) Write(p []byte) (n int, err error) {
	if len(w.m) == 0 {
		return w.w.Write(p)
	}
	out := mask.MaskSQLLogLine(string(p), w.m)
	_, err = w.w.Write([]byte(out))
	return len(p), err
}
