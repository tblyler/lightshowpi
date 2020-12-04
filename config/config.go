package config

import (
	"fmt"

	"github.com/pelletier/go-toml"
	"github.com/tblyler/lightshowpi/light/rpi"
	"github.com/tblyler/lightshowpi/sound"
)

// Config contains all the potentially needed config data
type Config struct {
	ListenAddress     string
	SongPath          string
	FMOutput          sound.FMConfig
	LightSchedulePath string
	RaspberryPi       struct {
		ChipName    string
		BasicLights map[string]struct {
			Pin int
		}
	}
}

// ParseConfig from raw binary data
func ParseConfig(rawConfig []byte) (*Config, error) {
	config := &Config{}

	err := toml.Unmarshal(rawConfig, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML config: %w", err)
	}

	return config, nil
}

// RPIManagerConfig instance derrived from this config
func (c *Config) RPIManagerConfig() *rpi.ManagerConfig {
	return &rpi.ManagerConfig{
		ChipName: c.RaspberryPi.ChipName,
	}
}
