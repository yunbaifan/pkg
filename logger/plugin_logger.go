package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	stringutil "github.com/yunbaifan/pkg/utils/strings"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type PluginLogger interface {
	CapitalColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder)
	CustomCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder)
}

const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

type Color uint8

func (c Color) Add(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}

var (
	_levelToColor = map[zapcore.Level]Color{
		zapcore.DebugLevel:  White,
		zapcore.InfoLevel:   Blue,
		zapcore.WarnLevel:   Yellow,
		zapcore.ErrorLevel:  Red,
		zapcore.DPanicLevel: Red,
		zapcore.PanicLevel:  Red,
		zapcore.FatalLevel:  Red,
	}
	Plugin *plugin
	_      PluginLogger = (*plugin)(nil)
	once   sync.Once
)

type plugin struct {
	toLowercaseColorString map[zapcore.Level]string
	toCapitalColorString   map[zapcore.Level]string
	unknownLevelColor      map[zapcore.Level]string
	modelName              string
	version                string
}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "gorm.io" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

var gormSourceDir string

func init() {
	Plugin = NewPlugin()
	for level, color := range _levelToColor {
		Plugin.toLowercaseColorString[level] = color.Add(level.String())
		Plugin.toCapitalColorString[level] = color.Add(level.CapitalString())
	}
	_, file, _, _ := runtime.Caller(0)

	gormSourceDir = sourceDir(file)
}

func NewPlugin() *plugin {
	once.Do(func() {
		Plugin = &plugin{
			toLowercaseColorString: make(map[zapcore.Level]string, len(_levelToColor)),
			toCapitalColorString:   make(map[zapcore.Level]string, len(_levelToColor)),
			unknownLevelColor:      make(map[zapcore.Level]string, len(_levelToColor)),
		}
	})
	return Plugin
}

func (p *plugin) SetNameAndVersion(name, version string) {
	p.modelName = name
	p.version = version
}

func (p *plugin) CapitalColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := p.toCapitalColorString[level]
	if !ok {
		s = p.unknownLevelColor[zapcore.ErrorLevel]
	}
	pid := stringutil.FormatString(fmt.Sprintf("["+"PID:"+"%d"+"]", os.Getpid), 15, true)
	color := _levelToColor[level]
	enc.AppendString(s)
	enc.AppendString(color.Add(s) + " " + pid)
	if p.modelName != "" {
		enc.AppendString(color.Add(stringutil.FormatString(p.modelName, 25, true)))
	}
	if p.version != "" {
		enc.AppendString(color.Add(
			stringutil.FormatString(
				fmt.Sprintf("["+"version:"+"%s"+"]",
					p.version,
				), 17, true)))
	}
}

func (p *plugin) CustomCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	fixedLength := 50
	trimmedPath := caller.TrimmedPath()
	trimmedPath = "[" + trimmedPath + "]"
	s := stringutil.FormatString(trimmedPath, fixedLength, true)
	enc.AppendString(s)
}

type alignEncoder struct {
	zapcore.Encoder
}

func (a *alignEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	entry.Message = fmt.Sprintf("%-50s", entry.Message)
	return a.Encoder.EncodeEntry(entry, fields)
}
