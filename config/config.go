package config

import "github.com/BurntSushi/toml"

// Config holds the configuration settings from the toml file.
type Config struct {
	DB Database `toml:"database"`
}

// Database holds the postgres database config.
type Database struct {
	Server string
	Port   int
}

// LoadConfig loads the given config toml file.
func LoadConfig(filepath string) (*Config, error) {
	c := &Config{}
	_, err := toml.DecodeFile(filepath, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
