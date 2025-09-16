package config

import (
	"github.com/spf13/viper"

	"github.com/strayca7/siam/internal/apiserver/options"
	"github.com/strayca7/siam/internal/pkg/util"
	namev1 "github.com/strayca7/siam/staging/src/api/name/v1"
)

// Load loads the configuration for the apiserver service from a YAML file.
// Load only can be called once during the application initialization.
func Load() (*options.Options, error) {
	v := viper.New()
	v.SetConfigName(namev1.APIServer)
	v.AddConfigPath(util.BaseConfigPath)
	v.SetConfigType(util.YAML)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	opts := options.NewOptions()
	if err := v.Unmarshal(opts); err != nil {
		return nil, err
	}
	return opts, nil
}
