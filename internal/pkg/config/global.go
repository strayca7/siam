package config

import (
	"github.com/spf13/viper"

	"github.com/strayca7/siam/internal/pkg/options"
	"github.com/strayca7/siam/internal/pkg/util"
)

// LoadGlobal uses viper to load global configuration file and returns the Global options.
// LoadGlobal only can be called once during the application initialization.
func LoadGlobal() (*options.Global, error) {
	v := viper.New()
	v.SetConfigName("global")
	v.AddConfigPath(util.BaseConfigPath)
	v.SetConfigType(util.YAML)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	opts := options.NewGlobal()
	if err := v.Unmarshal(opts); err != nil {
		return nil, err
	}
	return opts, nil
}
