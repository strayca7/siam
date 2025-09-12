package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

// W3C Trace Context (https://www.w3.org/TR/trace-context/)
// traceparent: "00-<trace-id>-<span-id>-<trace-flags>"
// trace-id: 16 bytes (32 hex) not all zeros
// span-id: 8 bytes (16 hex) not all zeros
// trace-flags: 1 byte (2 hex)

const (
	HeaderTraceParent = "traceparent"
	traceVersion      = "00"
)

// TraceContext represents the minimal W3C trace context we manage.
type TraceContext struct {
	Version    string
	TraceID    string
	SpanID     string
	ParentSpan string // optional, convenience (not part of traceparent itself; parent is previous span id)
	TraceFlags string // 2 hex chars, e.g. "01" sampled
}

// ParseTraceParent parses a traceparent header value.
func ParseTraceParent(v string) (TraceContext, error) {
	var tc TraceContext
	parts := strings.Split(v, "-")
	if len(parts) != 4 {
		return tc, errors.New("invalid traceparent parts")
	}
	ver, traceID, spanID, flags := parts[0], parts[1], parts[2], parts[3]
	if ver != traceVersion { // we accept only 00 for now
		return tc, errors.New("unsupported trace version")
	}
	if !validTraceID(traceID) {
		return tc, errors.New("invalid trace-id")
	}
	if !validSpanID(spanID) {
		return tc, errors.New("invalid span-id")
	}
	if len(flags) != 2 || !isHex(flags) {
		return tc, errors.New("invalid trace-flags")
	}
	tc = TraceContext{
		Version:    ver,
		TraceID:    strings.ToLower(traceID),
		SpanID:     strings.ToLower(spanID),
		TraceFlags: strings.ToLower(flags),
	}
	return tc, nil
}

// FormatTraceParent builds a traceparent header from TraceContext (ParentSpan ignored per spec — parent represented
// by previous span-id).
func FormatTraceParent(tc TraceContext) string {
	ver := tc.Version
	if ver == "" {
		ver = traceVersion
	}
	flags := tc.TraceFlags
	if flags == "" {
		flags = "01"
	} // default sampled
	return ver + "-" + tc.TraceID + "-" + tc.SpanID + "-" + flags
}

// validSpanID validates span-id (16 hex, not all zeros).
func validSpanID(id string) bool {
	if len(id) != 16 || !isHex(id) {
		return false
	}
	allZero := true
	for i := 0; i < 16; i++ {
		if id[i] != '0' {
			allZero = false
			break
		}
	}
	return !allZero
}

// isHex checks if string only hex digits.
func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// newSpanID returns 8 random bytes hex encoded.
func newSpanID() string {
	var b [8]byte
	for {
		if _, err := rand.Read(b[:]); err != nil {
			for i := range b {
				b[i] = byte(i + 11)
			}
		}
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

// ExtractTraceContext gets trace context from context (if present).
func ExtractTraceContext(ctx context.Context) TraceContext {
	tid, _ := ctx.Value(TraceIDKey).(string)
	sid, _ := ctx.Value(SpanIDKey).(string)
	psid, _ := ctx.Value(ParentSpanIDKey).(string)
	flags, _ := ctx.Value(TraceFlagsKey).(string)
	return TraceContext{Version: traceVersion, TraceID: tid, SpanID: sid, ParentSpan: psid, TraceFlags: flags}
}

// ContextWithTraceContext stores TraceContext in context.
func ContextWithTraceContext(ctx context.Context, tc TraceContext) context.Context {
	ctx = context.WithValue(ctx, TraceIDKey, tc.TraceID)
	ctx = context.WithValue(ctx, SpanIDKey, tc.SpanID)
	if tc.ParentSpan != "" {
		ctx = context.WithValue(ctx, ParentSpanIDKey, tc.ParentSpan)
	}
	if tc.TraceFlags != "" {
		ctx = context.WithValue(ctx, TraceFlagsKey, tc.TraceFlags)
	}
	return ctx
}

// StartSpan starts a new span: reuses existing trace-id, current span-id becomes parent, new span-id generated.
func StartSpan(ctx context.Context) (context.Context, TraceContext) {
	ctx, tid := EnsureTrace(ctx)
	curSpan, _ := ctx.Value(SpanIDKey).(string)
	newSpan := newSpanID()
	flags, _ := ctx.Value(TraceFlagsKey).(string)
	if flags == "" {
		flags = "01"
	}
	tc := TraceContext{Version: traceVersion, TraceID: tid, SpanID: newSpan, ParentSpan: curSpan, TraceFlags: flags}
	ctx = ContextWithTraceContext(ctx, tc)
	return ctx, tc
}

// InjectTraceParent sets the traceparent header into outgoing request.
func InjectTraceParent(ctx context.Context, h http.Header) {
	tc := ExtractTraceContext(ctx)
	if tc.TraceID == "" || tc.SpanID == "" {
		return
	}
	h.Set(HeaderTraceParent, FormatTraceParent(tc))
}

// WithIncomingRequest extracts traceparent from incoming headers; if absent creates root trace/span.
func WithIncomingRequest(ctx context.Context, h http.Header) (context.Context, TraceContext) {
	raw := h.Get(HeaderTraceParent)
	if raw != "" {
		if tc, err := ParseTraceParent(raw); err == nil {
			// existing span becomes parent when we start our own span
			ctx = ContextWithTraceContext(ctx, tc)
			// start a child span for our service scope
			return StartSpan(ctx)
		}
	}
	// create fresh root trace + span
	ctx, tid := EnsureTrace(ctx)
	span := newSpanID()
	tc := TraceContext{Version: traceVersion, TraceID: tid, SpanID: span, TraceFlags: "01"}
	ctx = ContextWithTraceContext(ctx, tc)
	return ctx, tc
}

// TraceID returns trace_id from context (helper, uses new keys).
func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(TraceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
