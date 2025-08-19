package logger

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/natefinch/lumberjack"
	"github.com/strayca7/siam/internal/pkg/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newOpts() *options.LoggerOptions {
	return &options.LoggerOptions{
		Name:       "siam",
		Level:      "info",
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
	}
}

var (
	env = os.Getenv("ENV")
	log *zap.Logger
	mu  sync.Mutex
)

type WithOpts func(*options.LoggerOptions)

func WithName(name string) WithOpts {
	return func(o *options.LoggerOptions) {
		o.Name = name
	}
}

func WithLevel(level string) WithOpts {
	return func(o *options.LoggerOptions) {
		o.Level = level
	}
}

func WithMaxSize(maxSize int) WithOpts {
	return func(o *options.LoggerOptions) {
		o.MaxSize = maxSize
	}
}

func WithMaxBackups(maxBackups int) WithOpts {
	return func(o *options.LoggerOptions) {
		o.MaxBackups = maxBackups
	}
}

func WithMaxAge(maxAge int) WithOpts {
	return func(o *options.LoggerOptions) {
		o.MaxAge = maxAge
	}
}

// Init initializes the logger with the given options.
// Then you can use logger.L() to get the logger instance.
func Init(opts *options.LoggerOptions, wo ...WithOpts) {
	o := newOpts()
	mu.Lock()
	defer mu.Unlock()
	if opts == nil {
		opts = o
	}
	for _, w := range wo {
		w(opts)
	}
	makeLogDir()
	new(opts)
}

func new(opts *options.LoggerOptions) {
	var core zapcore.Core

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(opts.Level)); err != nil {
		level = zap.InfoLevel
	}

	encCfg := zap.NewProductionEncoderConfig()
	
	if env == "dev" {
		encCfg.EncodeTime = zapcore.RFC3339TimeEncoder
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encCfg.EncodeCaller = zapcore.FullCallerEncoder
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encCfg)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join("log", opts.Name+".log"),
		MaxSize:    opts.MaxSize,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAge,
	})

	jsonEncoder := zapcore.NewJSONEncoder(encCfg)

	if env == "dev" {
		core = zapcore.NewTee(zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
			zapcore.NewCore(jsonEncoder, fileWriter, level))
		log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.DPanicLevel), zap.Fields(zap.String("svc", opts.Name)))
	} else {
		core = zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), level)
		log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.PanicLevel), zap.Fields(zap.String("svc", opts.Name)))
	}

	log.With(zap.Int("pid", os.Getpid()))
}

// L returns the logger instance.
// It must be called after Init.
func L() *zap.Logger {
	if log == nil {
		panic("logger not initialized")
	}
	return log
}

// S returns the sugared logger instance.
// It must be called after Init.
func S() *zap.SugaredLogger {
	return L().Sugar()
}

func makeLogDir() {
	if err := os.MkdirAll("./log", 0755); err != nil {
		log.Error("Failed to create log directory", zap.Error(err))
		panic(err)
	}
}
