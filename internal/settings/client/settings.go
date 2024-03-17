// Package client предоставляет настройки клиента.
package client

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

// Settings описывает структуру для хранения настроек клиента.
type Settings struct {
	Log       LogSettings      `koanf:"log"`
	Server    ServerSettings   `koanf:"server"`
	Asymmetry common.Asymmetry `koanf:"asymmetry"`
}

// LogSettings подструктура для хранения настроек логгера.
type LogSettings struct {
	Level   string `koanf:"level"`
	Verbose bool   `koanf:"verbose"`
	Format  string `koanf:"format"`
}

// ServerSettings подструктура для хранения настроек подключения к серверу.
type ServerSettings struct {
	Address string `koanf:"address"`
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
