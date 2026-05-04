package logx

import (
	"context"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/wisaitas/golang-structure/pkg/httpx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

type logger struct {
	zap   *zap.Logger
	level zap.AtomicLevel
}

func NewLogger(level string) Logger {
	parsed := parseLevel(level)
	atomic := zap.NewAtomicLevelAt(parsed)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.MessageKey = "message"
	encoderCfg.LevelKey = "level"
	encoderCfg.CallerKey = "caller"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(zapcore.AddSync(stdoutWriter{})),
		atomic,
	)

	zl := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return &logger{
		zap:   zl,
		level: atomic,
	}
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.write(ctx, zapcore.DebugLevel, msg, fields)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.write(ctx, zapcore.InfoLevel, msg, fields)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.write(ctx, zapcore.WarnLevel, msg, fields)
}

func (l *logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.write(ctx, zapcore.ErrorLevel, msg, fields)
}

func (l *logger) With(fields ...zap.Field) Logger {
	return &logger{
		zap:   l.zap.With(fields...),
		level: l.level,
	}
}

func (l *logger) Sync() error {
	return l.zap.Sync()
}

func (l *logger) write(ctx context.Context, lvl zapcore.Level, msg string, fields []zap.Field) {
	if !l.level.Enabled(lvl) {
		return
	}

	if collected := tryCollect(ctx, lvl, msg, fields); collected {
		return
	}

	switch lvl {
	case zapcore.DebugLevel:
		l.zap.Debug(msg, fields...)
	case zapcore.InfoLevel:
		l.zap.Info(msg, fields...)
	case zapcore.WarnLevel:
		l.zap.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		l.zap.Error(msg, fields...)
	default:
		l.zap.Info(msg, fields...)
	}
}

func tryCollect(ctx context.Context, lvl zapcore.Level, msg string, fields []zap.Field) bool {
	if ctx == nil {
		return false
	}

	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}

	caller := callerOutsideLogx()

	entry := httpx.AppLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     lvl.String(),
		Caller:    caller,
		Message:   msg,
		Fields:    enc.Fields,
	}

	return httpx.AddAppLog(ctx, entry)
}

func callerOutsideLogx() string {
	for skip := 3; skip < 10; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			return ""
		}
		if isLogxSourceFile(file) {
			continue
		}
		return filepath.Base(filepath.Dir(file)) + "/" + filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	return ""
}

func isLogxSourceFile(file string) bool {
	return strings.Contains(file, "/pkg/logx/") || strings.Contains(file, `\pkg\logx\`)
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel
	case "info", "":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
