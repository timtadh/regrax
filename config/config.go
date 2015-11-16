package config

import (
	"path/filepath"
)

import (
	"github.com/timtadh/sfp/stores/int_int"
	"github.com/timtadh/sfp/stores/ints_int"
	"github.com/timtadh/sfp/stores/ints_ints"
)


type Config struct {
	Cache string
	Output string
	Support, Samples, MinSize, MaxSize int
}

func (c *Config) CacheFile(name string) string {
	return filepath.Join(c.Cache, name)
}

func (c *Config) IntIntMultiMap(name string) (int_int.MultiMap, error) {
	if c.Cache == "" {
		return int_int.AnonBpTree()
	} else {
		return int_int.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

func (c *Config) IntsIntMultiMap(name string) (ints_int.MultiMap, error) {
	if c.Cache == "" {
		return ints_int.AnonBpTree()
	} else {
		return ints_int.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

func (c *Config) IntsIntsMultiMap(name string) (ints_ints.MultiMap, error) {
	if c.Cache == "" {
		return ints_ints.AnonBpTree()
	} else {
		return ints_ints.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

