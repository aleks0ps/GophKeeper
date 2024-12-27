// Package config -- описывает настройки сервиса
package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	defaultRunAddress = "localhost:8080"
)

type Config struct {
	RunAddress  string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
	Secret      string `env:"SECRET_KEY"`
}

func ParseOptions() *Config {
	opts := Config{
		RunAddress: defaultRunAddress,
	}
	if err := env.Parse(&opts); err != nil {
		fmt.Println("failed:", err)
	}
	flag.StringVar(&opts.RunAddress, "l", opts.RunAddress, "Listen address:port")
	flag.StringVar(&opts.DatabaseURI, "d", opts.DatabaseURI, "Postgres connection string")
	flag.StringVar(&opts.Secret, "s", opts.Secret, "Secret key for data encryption")
	flag.Parse()
	return &opts
}
