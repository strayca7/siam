package options

import (
	"log"

	"github.com/spf13/viper"

	"github.com/strayca7/siam/internal/pkg/util"
)

// Global holds the configuration values for the application.
// Each service must initialize its own configuration.
type Global struct {
	Log *Logger `json:"log" mapstructure:"log"`
}

func NewGlobal() *Global {
	viper.SetConfigName("global")
	viper.AddConfigPath(util.BaseConfigPath)
	viper.SetConfigType("yaml")

	var global Global

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&global); err != nil {
		log.Fatal(err)
	}
	return &global
}
