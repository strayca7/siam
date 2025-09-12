package options

type Logger struct {
	Name       string `json:"name"       mapstructure:"name"`
	Level      string `json:"level"      mapstructure:"level"`
	MaxSize    int    `json:"maxSize"    mapstructure:"maxSize"`
	MaxBackups int    `json:"maxBackups" mapstructure:"maxBackups"`
	MaxAge     int    `json:"maxAge"     mapstructure:"maxAge"`

	// If true, enable request traceID and spanID logging
	EnableTrace bool `json:"enableTrace" mapstructure:"enableTrace"`
}

// NewLogger creates a new Logger instance with the specified name.
func NewLogger() *Logger {
	return &Logger{
		Name:        "siam",
		Level:       "info",
		MaxSize:     10, // megabytes
		MaxBackups:  5,
		MaxAge:      30, // days
		EnableTrace: false,
	}
}
