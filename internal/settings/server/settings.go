// Package server предоставляет настройки сервера.
package server

import (
	"GophKeeper/internal/settings/common"
	"fmt"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"

	"github.com/knadh/koanf/v2"
)

const (
	defaultConfigFile = "config.yaml"
)

// Settings описывает структуру для хранения настроек сервера.
type Settings struct {
	API       APISettings      `koanf:"api"`
	Database  DatabaseSettings `koanf:"database"`
	Log       LogSettings      `koanf:"log"`
	JWT       JwtSettings      `koanf:"jwt"`
	Crypto    CryptoSettings   `koanf:"crypto"`
	Asymmetry common.Asymmetry `koanf:"asymmetry"`
}

// APISettings подструктура для хранения настроек API.
type APISettings struct {
	Address string `koanf:"address"`
	Port    int    `koanf:"port"`
}

// DatabaseSettings подструктура для хранения настроек подключения к БД.
type DatabaseSettings struct {
	URI string `koanf:"uri"`
}

// LogSettings подструктура для хранения настроек логгера.
type LogSettings struct {
	Level   string `koanf:"level"`
	Verbose bool   `koanf:"verbose"`
	Format  string `koanf:"format"`
}

// JwtSettings подструктура для хранения настроек для jwt-токена.
type JwtSettings struct {
	Secret   string `koanf:"secret"`
	Lifetime struct {
		Access  string `koanf:"access"`
		Refresh string `koanf:"refresh"`
	} `koanf:"lifetime"`
}

// CryptoSettings подструктура для хранения настроек шифрования.
type CryptoSettings struct {
	Key  string `koanf:"key"`
	Salt string `koanf:"salt"`
}

// NewSettings принимает путь до файла настроек и пытается создать объект Settings.
func NewSettings(config string) (*Settings, error) {
	if config == "" {
		config = defaultConfigFile
	}

	k := koanf.New(".")

	// Read configuration file if exists.
	err := k.Load(file.Provider(config), yaml.Parser())
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", config, err)
	}

	s := &Settings{}

	err = k.Unmarshal("", &s)
	if err != nil {
		return nil, fmt.Errorf("unmarshal configuration: %w", err)
	}

	return s, nil
}
