package config

import (
	"path/filepath"
)

import (
	"github.com/timtadh/sfp/stores/intint"
	"github.com/timtadh/sfp/stores/intsints"
	"github.com/timtadh/sfp/stores/itemset_int"
	"github.com/timtadh/sfp/stores/itemsets"
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

func (c *Config) IntsIntsMultiMap(name string) (intsints.MultiMap, error) {
	if c.Cache == "" {
		return intsints.AnonBpTree()
	} else {
		return intsints.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

func (c *Config) ItemsetIntMultiMap(name string) (itemset_int.MultiMap, error) {
	if c.Cache == "" {
		return itemset_int.AnonBpTree()
	} else {
		return itemset_int.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

func (c *Config) ItemsetsMultiMap(name string) (itemsets.MultiMap, error) {
	if c.Cache == "" {
		return itemsets.AnonBpTree()
	} else {
		return itemsets.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

