package options

// Global holds the configuration values for the application.
// Each service must initialize its own configuration.
type Global struct {
	Log *Logger `json:"log" mapstructure:"log"`
}

func NewGlobal() *Global {
	return &Global{
		Log: NewLogger(),
	}
}
