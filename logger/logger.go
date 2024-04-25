package logger

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/yunbaifan/pkg/imcontext"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log interface {
	Debug(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, err error, fields ...any)
	Error(ctx context.Context, msg string, err error, fields ...any)
	WithValues(fields ...any) Log
	WithName(name string) Log
	WithDepth(depth int) Log
}

var (
	_      Log = (*zapLogger)(nil)
	Logger Log
	mu     sync.Mutex
	sp     = string(filepath.Separator)
)

type LoggerConifg struct {
	LogPrefixName string        `json:"logPrefixName" yaml:"logPrefixName" default:"go-mall" description:"日志文件前缀名"`
	ModuleName    string        `json:"moduleName" yaml:"moduleName" default:"go-mall" description:"模块名称"`
	LogLevel      int           `json:"logLevel" yaml:"logLevel" default:"info" description:"日志级别"`
	IsStdout      bool          `json:"isStdout" yaml:"isStdout" default:"true" description:"是否输出到终端"`
	IsJson        bool          `json:"isJson" yaml:"isJson" default:"true" description:"是否输出为json格式"`
	Location      string        `json:"logLocation" yaml:"logLocation" default:"logs/*.log" description:"日志文件路径"`
	RotateCount   uint          `json:"rotateCount" yaml:"rotateCount" default:"7" description:"日志文件最大保存天数"`
	RotationTime  time.Duration `json:"rotationTime" yaml:"rotationTime" default:"24" description:"日志文件最大保存天数"`
	Version       string        `json:"version" yaml:"version" default:"v1.0.0" description:"版本号"`
	PId           int           `json:"pid" yaml:"pid" default:"0" description:"进程ID"`
}

type zapLogger struct {
	zap          *zap.SugaredLogger
	level        zapcore.Level
	name         string
	preName      string
	rotationTime time.Duration
	layout       string
	PId          int
	version      string
	plugin       PluginLogger
}

func getLevel(level int) zapcore.Level {
	maps := map[int]zapcore.Level{
		6: zapcore.DebugLevel,
		5: zapcore.DebugLevel,
		4: zapcore.InfoLevel,
		3: zapcore.WarnLevel,
		2: zapcore.ErrorLevel,
		1: zapcore.FatalLevel,
		0: zapcore.PanicLevel,
	}
	return maps[level]
}

func NewZapLogger(cfg *LoggerConifg) (Log, error) {
	if Logger == nil {
		mu.Lock()
		defer mu.Unlock()
		if Logger == nil {
			zapConfig := zap.Config{
				Level:             zap.NewAtomicLevelAt(getLevel(cfg.LogLevel)),
				DisableStacktrace: true,
			}
			if cfg.IsJson {
				zapConfig.Encoding = "json"
			} else {
				zapConfig.Encoding = "console"
			}
			zl := &zapLogger{
				level:        getLevel(cfg.LogLevel),
				name:         cfg.ModuleName,
				version:      cfg.Version,
				preName:      cfg.LogPrefixName,
				rotationTime: cfg.RotationTime * time.Hour,
				layout:       "2006-01-02 15:04:05",
				PId:          cfg.PId,
				plugin:       NewPlugin(),
			}
			opts, err := zl.core(cfg.IsStdout, cfg.IsJson, cfg.Location, cfg.RotateCount)
			if err != nil {
				return nil, err
			}
			l, err := zapConfig.Build(opts)
			if err != nil {
				return nil, err
			}
			zl.zap = l.Sugar()
			Logger = zl
		}
	}
	return Logger, nil
}

func (z *zapLogger) timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, z.layout)
		return
	}
	enc.AppendString(t.Format(z.layout))

}

func (z *zapLogger) core(isStdout, isJson bool, logLoction string, rotateCount uint) (zap.Option, error) {
	c := zap.NewProductionEncoderConfig()
	c.EncodeTime = z.timeEncoder
	c.EncodeDuration = zapcore.SecondsDurationEncoder
	c.MessageKey = "msg"
	c.LevelKey = "level"
	c.TimeKey = "time"
	c.NameKey = "logger"
	var fileEncoder zapcore.Encoder
	if isJson {
		c.EncodeLevel = zapcore.CapitalLevelEncoder
		fileEncoder = zapcore.NewJSONEncoder(c)
		fileEncoder.AddInt("PID", os.Getegid())
		fileEncoder.AddString("version", z.version)
	} else {
		c.EncodeLevel = z.plugin.CapitalColorLevelEncoder
		c.EncodeCaller = z.plugin.CustomCallerEncoder
		fileEncoder = zapcore.NewConsoleEncoder(c)
	}
	fileEncoder = &alignEncoder{fileEncoder}
	writer, err := z.getWriter(logLoction, rotateCount)
	if err != nil {
		return nil, err
	}
	var (
		cores []zapcore.Core
	)
	if logLoction != "" {
		cores = []zapcore.Core{
			zapcore.NewCore(fileEncoder, writer, zap.NewAtomicLevelAt(z.level)),
		}
	}
	if isStdout {
		cores = append(cores,
			zapcore.NewCore(fileEncoder, zapcore.Lock(os.Stdout),
				zap.NewAtomicLevelAt(z.level),
			))
	}
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	}), nil
}

func (z *zapLogger) getWriter(logLocation string, rotateCount uint) (zapcore.WriteSyncer, error) {
	var path string
	if z.rotationTime%(time.Hour*time.Duration(24)) == 0 {

	}
	switch {
	case z.rotationTime%(time.Hour*time.Duration(24)) == 0:
		path = logLocation + sp + z.preName + ".%Y-%m-%d.log"
	case z.rotationTime%time.Hour == 0:
		path = logLocation + sp + z.preName + ".%Y-%m-%d_%H"
	default:
		path = logLocation + sp + z.preName + ".%Y-%m-%d_%H-%M-%S"
	}
	logf, err := rotatelogs.New(
		path,
		rotatelogs.WithRotationCount(rotateCount),
		rotatelogs.WithRotationTime(z.rotationTime),
	)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(logf), nil
}

func (z *zapLogger) Debug(ctx context.Context, msg string, fields ...any) {
	if z.level > zapcore.DebugLevel {
		return
	}
	kv := z.AppendString(ctx, fields)
	z.zap.Debugw(msg, kv...)
}

func (z *zapLogger) Info(ctx context.Context, msg string, fields ...any) {
	if z.level > zapcore.InfoLevel {
		return
	}
	kv := z.AppendString(ctx, fields)
	z.zap.Infow(msg, kv...)
}

func (z *zapLogger) Warn(ctx context.Context, msg string, err error, fields ...any) {
	if z.level > zapcore.WarnLevel {
		return
	}
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	kv := z.AppendString(ctx, fields)
	z.zap.Warnw(msg, kv...)
}

func (z *zapLogger) Error(ctx context.Context, msg string, err error, fields ...any) {
	if z.level > zapcore.ErrorLevel {
		return
	}
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	kv := z.AppendString(ctx, fields)
	z.zap.Errorw(msg, kv...)
}

func (z *zapLogger) AppendString(ctx context.Context, kv []any) []any {
	if operationID := imcontext.GetOperation(ctx); operationID != "" {
		kv = append([]any{"operationID", operationID}, kv...)
	}
	if opUserID := imcontext.GetOpUserID(ctx); opUserID != "" {
		kv = append([]any{"opUserID", opUserID}, kv...)
	}
	if opUserPlatform := imcontext.GetOpUserPlatform(ctx); opUserPlatform != "" {
		kv = append([]any{"opUserPlatform", opUserPlatform}, kv...)
	}
	if connID := imcontext.GetConnID(ctx); connID != "" {
		kv = append([]any{"connID", connID}, kv...)
	}
	if triggerID := imcontext.GetTriggerID(ctx); triggerID != "" {
		kv = append([]any{"triggerID", triggerID}, kv...)
	}
	if remoteAddr := imcontext.GetRemoteAddr(ctx); remoteAddr != "" {
		kv = append([]any{"remoteAddr", remoteAddr}, kv...)
	}
	return kv
}

func (z *zapLogger) WithValues(fields ...any) Log {
	dup := *z
	dup.zap = z.zap.With(fields...)
	return &dup
}
func (l *zapLogger) WithName(name string) Log {
	dup := *l
	dup.zap = l.zap.Named(name)
	return &dup
}

func (l *zapLogger) WithDepth(depth int) Log {
	dup := *l
	dup.zap = l.zap.WithOptions(zap.AddCallerSkip(depth))
	return &dup
}
