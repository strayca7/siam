package options

import "github.com/strayca7/siam/internal/pkg/options"

type Options struct {
	Postgres options.Postgres `json:"postgres" mapstructure:"postgres"`
}

func NewOptions() *Options {
	return &Options{
		Postgres: *options.NewPostgres(),
	}
}
