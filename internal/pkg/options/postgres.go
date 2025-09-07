package options

import (
	"github.com/strayca7/siam/pkg/db"
	"gorm.io/gorm"
)

type Postgres struct {
	Host            string `json:"host" mapstructure:"host"`
	User            string `json:"user" mapstructure:"user"`
	Password        string `json:"password" mapstructure:"password"`
	Datebase        string `json:"database" mapstructure:"database"`
	Port            int    `json:"port" mapstructure:"port"`
	SSLMode         string `json:"sslMode" mapstructure:"sslMode"`
	TimeZone        string `json:"timeZone" mapstructure:"timeZone"`
	MaxIdleConns    int    `json:"maxIdleConns" mapstructure:"maxIdleConns"`
	MaxOpenConns    int    `json:"maxOpenConns" mapstructure:"maxOpenConns"`
	ConnMaxIdleTime int    `json:"connMaxIdleTime" mapstructure:"connMaxIdleTime"`
	ConnMaxLifetime int    `json:"connMaxLifetime" mapstructure:"connMaxLifetime"`
}

// NewPostgres creates a `zero` value instance.
func NewPostgres() *Postgres {
	return &Postgres{
		Port:            5432,
		User:            "",
		Password:        "",
		Datebase:        "",
		SSLMode:         "disable",
		TimeZone:        "Asia/Shanghai",
		MaxIdleConns:    100,
		MaxOpenConns:    100,
		ConnMaxIdleTime: 10,
		ConnMaxLifetime: 30,
	}
}

// NewPostgresCli creates a new gorm db instance with the given options.
// This logic is waiting to split into options and db package.
func (o *Postgres) NewPostgresCli() (*gorm.DB, error) {
	opts := &db.Options{
		Host:            o.Host,
		User:            o.User,
		Password:        o.Password,
		Datebase:        o.Datebase,
		Port:            o.Port,
		SSLMode:         o.SSLMode,
		TimeZone:        o.TimeZone,
		MaxIdleConns:    o.MaxIdleConns,
		MaxOpenConns:    o.MaxOpenConns,
		ConnMaxIdleTime: o.ConnMaxIdleTime,
		ConnMaxLifetime: o.ConnMaxLifetime,
	}
	return db.New(opts)
}
