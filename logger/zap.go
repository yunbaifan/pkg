package logger

import (
	"os"
	"sync"
	"time"

	"context"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	LogConfig struct {
		Level      string `json:"level" yaml:"level" default:"info" description:"日志级别"`
		Filename   string `json:"fileName" yaml:"fileName" default:"logs/go-mall.log" description:"日志文件路径"`
		MaxSize    int    `json:"maxSize" yaml:"maxSize" default:"100" description:"日志文件最大大小(MB)"`
		MaxAge     int    `json:"maxAge" yaml:"maxAge" default:"7" description:"日志文件最大保存天数"`
		MaxBackups int    `json:"maxBackups" yaml:"maxBackups" default:"10" description:"日志文件最多保存多少个备份"`
	}
)

var (
	syncOnce sync.Once
	Logger   *zap.Logger
)

// InitLogger 初始化Logger
func InitLogger(cfg LogConfig) (err error) {
	syncOnce.Do(func() {
		writeSyncer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
		encoder := getEncoder()
		var l = new(zapcore.Level)
		err = l.UnmarshalText([]byte(cfg.Level))
		if err != nil {
			return
		}
		var core zapcore.Core
		if cfg.Level == "debug" {
			// 进入开发模式，日志输出到终端
			config := zap.NewDevelopmentEncoderConfig()
			// 设置日志颜色
			config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
			// 设置自定义时间格式
			config.EncodeTime = getCustomTimeEncoder
			consoleEncoder := zapcore.NewConsoleEncoder(config)
			core = zapcore.NewTee(
				zapcore.NewCore(encoder, writeSyncer, l),
				zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
			)
		} else {
			core = zapcore.NewCore(encoder, writeSyncer, l)
		}

		Logger = zap.New(core, zap.AddCaller())

		zap.ReplaceGlobals(Logger)
		Logger.Info("init logger success")

	})
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// CustomTimeEncoder 自定义日志输出时间格式
func getCustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[go-mall]" + t.Format("2006/01/02 - 15:04:05.000"))
}

func LogWith(withs ...zap.Field) *zap.Logger {
	return Logger.With(withs...)
}

type logCtxKey struct{}

func ZapLoggerContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, logCtxKey{}, logger)
}

func FromZapLoggerContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(logCtxKey{}).(*zap.Logger); ok {
		return logger
	}
	return Logger
}
