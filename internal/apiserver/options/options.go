package options

import genericoptions "github.com/strayca7/siam/internal/pkg/options"

type Options struct {
	Postgres *genericoptions.Postgres `json:"postgres" mapstructure:"postgres"`
}

func NewOptions() *Options {
	return &Options{
		Postgres: genericoptions.NewPostgres(),
	}
}
