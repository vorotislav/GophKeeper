package server

import (
	"fmt"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"

	"github.com/knadh/koanf/v2"
)

type Settings struct {
	API      APISettings      `koanf:"api"`
	Database DatabaseSettings `koanf:"database"`
	Log      LogSettings      `koanf:"log"`
	JWT      JwtSettings      `koanf:"jwt"`
	Crypto   CryptoSettings   `koanf:"crypto"`
}

type APISettings struct {
	Address string `koanf:"address"`
	Port    int    `koanf:"port"`
}

type DatabaseSettings struct {
	URI string `koanf:"uri"`
}

type LogSettings struct {
	Level   string `koanf:"level"`
	Verbose bool   `koanf:"verbose"`
	Format  string `koanf:"format"`
}

type JwtSettings struct {
	Secret   string `koanf:"secret"`
	Lifetime struct {
		Access  string `koanf:"access"`
		Refresh string `koanf:"refresh"`
	} `koanf:"lifetime"`
}

type CryptoSettings struct {
	Key  string `koanf:"key"`
	Salt string `koanf:"salt"`
}

func NewSettings(config string) (*Settings, error) {
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
