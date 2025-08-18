package options

// Global holds the configuration values for the application.
// It satisfies all sub config interfaces like LogConfig.
type Global struct {
	Log *LoggerOptions `yaml:"log"`
}
