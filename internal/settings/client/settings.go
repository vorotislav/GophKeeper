package client

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Settings struct {
	Log    LogSettings    `koanf:"log"`
	Server ServerSettings `koanf:"server"`
}

type LogSettings struct {
	Level   string `koanf:"level"`
	Verbose bool   `koanf:"verbose"`
	Format  string `koanf:"format"`
}

type ServerSettings struct {
	Address string `koanf:"address"`
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
