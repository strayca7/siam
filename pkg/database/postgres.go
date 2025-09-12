package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresOptions defines the configuration options for the Postgres database.
type PostgresOptions struct {
	Host            string
	User            string
	Password        string
	Database        string
	Port            int
	SSLMode         string
	TimeZone        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxIdleTime int
	ConnMaxLifetime int
}

// // New create a new gorm db instance with the given options.
func New(opts *PostgresOptions) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		opts.Host,
		opts.User,
		opts.Password,
		opts.Database,
		opts.Port,
		opts.SSLMode,
		opts.TimeZone,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqldb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqldb.SetMaxIdleConns(opts.MaxIdleConns)
	sqldb.SetMaxOpenConns(opts.MaxOpenConns)
	sqldb.SetConnMaxIdleTime(time.Duration(opts.ConnMaxIdleTime) * time.Minute)
	sqldb.SetConnMaxLifetime(time.Duration(opts.ConnMaxLifetime) * time.Minute)
	return db, nil
}
