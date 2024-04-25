package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gormLogger "gorm.io/gorm/logger"
	gormUtils "gorm.io/gorm/utils"
)

type ormLogger struct {
	logger.Config
	nanosecondsToMilliseconds float64
}

type Option func(c *ormLogger)

func WithSlowThreshold(threshold time.Duration) Option {
	return func(c *ormLogger) {
		c.SlowThreshold = threshold * time.Millisecond
	}
}

func WithLogLevel(level logger.LogLevel) Option {
	return func(c *ormLogger) {
		c.LogLevel = level
	}
}
func WithColorful(colorful bool) Option {
	return func(c *ormLogger) {
		c.Colorful = colorful
	}
}

func NewGormLogger(opt ...Option) logger.Interface {
	newLogger := &ormLogger{
		Config: logger.Config{
			SlowThreshold: 100 * time.Millisecond,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
		nanosecondsToMilliseconds: 1e6,
	}
	for _, o := range opt {
		o(newLogger)
	}
	return newLogger
}

// LogMode log mode
func (l *ormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info print info
func (l ormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	Info(ctx, msg, data...)
}

// Warn print warn messages
func (l ormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	Warn(ctx, msg, nil, data...)
}

// Error print error messages
func (l ormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	Error(ctx, msg, nil, data...)
}

func (l ormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormLogger.Error &&
		(!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			Error(ctx, "sql exec detail", err, "gorm",
				gormUtils.FileWithLineNum(), "elapsed time",
				fmt.Sprintf("%f(ms)", float64(elapsed.Nanoseconds())/l.nanosecondsToMilliseconds),
				"sql", sql)
		} else {
			Error(ctx, "sql exec detail", err, "gorm",
				gormUtils.FileWithLineNum(), "elapsed time",
				fmt.Sprintf("%f(ms)", float64(elapsed.Nanoseconds())/l.nanosecondsToMilliseconds),
				"rows", rows, "sql", sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			Warn(ctx, "sql exec detail", nil, "gorm",
				gormUtils.FileWithLineNum(), "slow sql",
				slowLog, "elapsed time",
				fmt.Sprintf("%f(ms)",
					float64(elapsed.Nanoseconds())/l.nanosecondsToMilliseconds), "sql", sql)
		} else {
			Warn(ctx, "sql exec detail", nil, "gorm",
				gormUtils.FileWithLineNum(), "slow sql",
				slowLog, "elapsed time",
				fmt.Sprintf("%f(ms)", float64(elapsed.Nanoseconds())/l.nanosecondsToMilliseconds),
				"rows", rows, "sql", sql)
		}
	case l.LogLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			Debug(ctx, "sql exec detail", "gorm",
				gormUtils.FileWithLineNum(), "elapsed time", fmt.Sprintf("%f(ms)",
					float64(elapsed.Nanoseconds())/l.nanosecondsToMilliseconds), "sql", sql)
		} else {
			Debug(ctx, "sql exec detail", "gorm",
				gormUtils.FileWithLineNum(), "elapsed time",
				fmt.Sprintf("%f(ms)",
					float64(elapsed.Nanoseconds())/l.nanosecondsToMilliseconds),
				"rows", rows, "sql", sql)
		}
	}
}
