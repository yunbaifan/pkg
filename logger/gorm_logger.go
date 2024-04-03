package logger

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type ormLogger struct {
	logger.Config
}

type Option func(c *ormLogger)

func WithSlowThreshold(threshold time.Duration) Option {
	return func(c *ormLogger) {
		c.SlowThreshold = threshold
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
		logger.Config{
			SlowThreshold: 100 * time.Millisecond,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
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
	if l.LogLevel >= logger.Info {
		logx.WithContext(ctx).Infow(
			msg,
			logx.Field("file", utils.FileWithLineNum()),
			logx.Field("rawData", data),
		)
	}
}

// Warn print warn messages
func (l ormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		logx.WithContext(ctx).Infow(
			msg,
			logx.Field("file", utils.FileWithLineNum()),
			logx.Field("rawData", data),
		)
	}
}

// Error print error messages
func (l ormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		logx.WithContext(ctx).Errorw(
			msg,
			logx.Field("file", utils.FileWithLineNum()),
			logx.Field("rawData", data),
		)
	}
}

func (l ormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error:
		sql, rows := fc()
		logx.WithContext(ctx).Errorw(
			"gorm trace log",
			logx.Field("file", utils.FileWithLineNum()),
			logx.Field("err", err),
			logx.Field("elapsed", float64(elapsed.Nanoseconds())/1e6),
			logx.Field("rows", rows),
			logx.Field("sql", sql),
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		logx.WithContext(ctx).Errorw(
			"gorm trace log",
			logx.Field("file", utils.FileWithLineNum()),
			logx.Field("slowLog", fmt.Sprintf("slow sql more than %v", l.SlowThreshold)),
			logx.Field("elapsed", float64(elapsed.Nanoseconds())/1e6),
			logx.Field("rows", rows),
			logx.Field("sql", sql),
		)
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		logx.WithContext(ctx).Infow(
			"gorm info log",
			logx.Field("file", utils.FileWithLineNum()),
			logx.Field("elapsed", float64(elapsed.Nanoseconds())/1e6),
			logx.Field("rows", rows),
			logx.Field("sql", sql),
		)
	}
}
