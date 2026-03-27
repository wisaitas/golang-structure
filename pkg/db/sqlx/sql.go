package sqlx

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/mask"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
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
	maskMap := mask.ParsePatternMap(config.MaskPattern)
	gcfg.Logger = &collectDBLogger{
		maskMap:  maskMap,
		logLevel: logger.Warn,
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

type collectDBLogger struct {
	maskMap  map[string]string
	logLevel logger.LogLevel
}

func (l *collectDBLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &collectDBLogger{maskMap: l.maskMap, logLevel: level}
}

func (l *collectDBLogger) Info(ctx context.Context, msg string, data ...interface{}) {
}

func (l *collectDBLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
}

func (l *collectDBLogger) Error(ctx context.Context, msg string, data ...interface{}) {
}

func (l *collectDBLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel == logger.Silent {
		return
	}

	sql, rows := fc()
	if len(l.maskMap) > 0 {
		sql = mask.MaskSQLLogLine(sql, l.maskMap)
	}

	var errMsg *string
	if err != nil {
		msg := err.Error()
		errMsg = &msg
	}

	httpx.AddDBLog(ctx, httpx.DBLog{
		Source:     normalizeSourcePath(utils.FileWithLineNum()),
		SQL:        sql,
		Rows:       rows,
		DurationMs: time.Since(begin).Milliseconds(),
		Error:      errMsg,
	})
}

func normalizeSourcePath(source string) string {
	const marker = "/golang-structure/"
	idx := strings.Index(source, marker)
	if idx == -1 {
		return source
	}
	return source[idx+1:]
}
