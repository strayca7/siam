package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/strayca7/siam/internal/pkg/options"
)

func newOpts() *options.Logger {
	return &options.Logger{
		Name:        "siam",
		Level:       "info",
		MaxSize:     10, // megabytes
		MaxBackups:  5,
		MaxAge:      30, // days
		EnableTrace: false,
	}
}

var (
	env = os.Getenv("ENV")
	log *zap.Logger
	mu  sync.Mutex
)

type WithOpts func(*options.Logger)

func WithName(name string) WithOpts {
	return func(o *options.Logger) {
		o.Name = name
	}
}

func WithLevel(level string) WithOpts {
	return func(o *options.Logger) {
		o.Level = level
	}
}

func WithMaxSize(maxSize int) WithOpts {
	return func(o *options.Logger) {
		o.MaxSize = maxSize
	}
}

func WithMaxBackups(maxBackups int) WithOpts {
	return func(o *options.Logger) {
		o.MaxBackups = maxBackups
	}
}

func WithMaxAge(maxAge int) WithOpts {
	return func(o *options.Logger) {
		o.MaxAge = maxAge
	}
}

// Init initializes the global logger with Logger and the given options.
// Its first argument is the context.Context, which is used to extract trace_id, span_id and svc_id which format in W3C.
// If you use an empty Logger, the default options will be used.
// Then you can use logger.L() to get the logger instance.
func Init(ctx context.Context, opts *options.Logger, wo ...WithOpts) {
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
	log = new(ctx, opts)
}

// New returns a new initialized logger with the given options.
func New(ctx context.Context, opts *options.Logger, wo ...WithOpts) *zap.Logger {
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
	return new(ctx, opts)
}

// Context keys (string kept for backward compatibility; prefer unexported types in new code)
const (
	TraceIDKey      = "trace_id" // 128-bit (16 byte) lowercase hex string per W3C Trace Context trace-id format
	SvcIDKey        = "svc_id"
	SpanIDKey       = "span_id"
	ParentSpanIDKey = "parent_span_id"
	TraceFlagsKey   = "trace_flags"
)

// ContextWithTrace returns a new context carrying the provided traceID (must be 32 hex chars).
func ContextWithTrace(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// EnsureTrace returns a context that has a trace_id (generates one if missing) and the trace id string.
// It follows W3C trace-id requirements: 16 bytes, all zero not allowed, represented as 32 lowercase hex chars.
func EnsureTrace(ctx context.Context) (context.Context, string) {
	if ctx == nil {
		ctx = context.Background()
	}
	if tid, _ := ctx.Value(TraceIDKey).(string); validTraceID(tid) {
		return ctx, tid
	}
	tid := newTraceID()
	ctx = context.WithValue(ctx, TraceIDKey, tid)
	return ctx, tid
}

// newTraceID generates a 16-byte (128-bit) random trace id encoded as 32 lowercase hex.
func newTraceID() string {
	var b [16]byte
	for {
		if _, err := rand.Read(b[:]); err != nil {
			// fallback: should rarely happen; use zeros replaced with time-based random-ish sequence not imported to
			// keep minimal deps
			for i := range b {
				b[i] = byte(i + 1)
			}
		}
		// must be not all zero per spec
		allZero := true
		for _, v := range b {
			if v != 0 {
				allZero = false
				break
			}
		}
		if !allZero {
			break
		}
	}
	return hex.EncodeToString(b[:])
}

// validTraceID performs a minimal validation: length 32 hex chars and not all zeros.
func validTraceID(id string) bool {
	if len(id) != 32 {
		return false
	}
	zero := true
	for i := 0; i < 32; i++ {
		c := id[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
		if c != '0' {
			zero = false
		}
	}
	return !zero
}

func new(ctx context.Context, opts *options.Logger) *zap.Logger {
	// Ensure trace id and span id exist (root span if missing)
	ctx, traceID := EnsureTrace(ctx)
	spanID, _ := ctx.Value(SpanIDKey).(string)
	if !validSpanID(spanID) { // create a root span
		spanID = newSpanID()
		ctx = ContextWithTraceContext(
			ctx,
			TraceContext{Version: traceVersion, TraceID: traceID, SpanID: spanID, TraceFlags: "01"},
		)
	}
	svcID, _ := ctx.Value(SvcIDKey).(string)

	var core zapcore.Core

	// if opts.Level is invalid, panic
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(opts.Level)); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level %q: %v\n", opts.Level, err)
		panic(err)
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

	var log *zap.Logger
	if env == "dev" {
		core = zapcore.NewTee(zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
			zapcore.NewCore(jsonEncoder, fileWriter, level))
		log = zap.New(
			core,
			zap.AddCaller(),
			zap.AddCallerSkip(0),
			zap.AddStacktrace(zap.DPanicLevel),
			zap.Fields(zap.String("svc", opts.Name)),
		)
	} else {
		core = zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), level)
		log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0), zap.AddStacktrace(zap.PanicLevel), zap.Fields(zap.String("svc", opts.Name)))
	}

	if opts.EnableTrace {
		return log.With(
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
			zap.String("svc_id", svcID),
		)
	}

	if svcID == "" {
		return log
	}
	return log.With(zap.String("svc_id", svcID))
}

// L returns the logger instance.
// It must be called after Init().
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
