package options

type LoggerOptions struct {
	Name       string `yaml:"name"`
	Level      string `yaml:"level"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
}
