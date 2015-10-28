package config

import (
	"path/filepath"
)


type Config struct {
	Cache string
	Output string
}

func (c *Config) CacheFile(name string) string {
	return filepath.Join(c.Cache, name)
}

