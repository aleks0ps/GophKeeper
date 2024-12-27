// Package config -- пакет описывает конфигурацию клиента
package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	defaultURL      = "https://localhost:8080"
	defaultDownload = "download"
)

// Config -- структура содержит значение параметров приложения
type Config struct {
	URL      string `env:"SERVER_URL" json:"server_url"`
	Download string `env:"DOWNLOAD_DIR" json:"download_dir"`
}

// ParseOptions -- создает конфиг приложения из переданных пользователем опций
func ParseOptions() *Config {
	opts := Config{URL: defaultURL, Download: defaultDownload}
	// Read json config
	if err := env.Parse(&opts); err != nil {
		fmt.Println("failed:", err)
	}
	flag.StringVar(&opts.URL, "h", opts.URL, "URL to connect to")
	flag.StringVar(&opts.Download, "d", opts.Download, "Download dir")
	flag.Parse()
	return &opts
}
