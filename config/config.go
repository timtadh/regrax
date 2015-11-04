package config

import (
	"path/filepath"
)

import (
	"github.com/timtadh/sfp/stores/intint"
)


type Config struct {
	Cache string
	Output string
	Support, Samples, MinSize, MaxSize int
}

func (c *Config) CacheFile(name string) string {
	return filepath.Join(c.Cache, name)
}

func (c *Config) IntIntMultiMap(name string) (intint.MultiMap, error) {
	if c.Cache == "" {
		return intint.AnonBpTree()
	} else {
		return intint.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

