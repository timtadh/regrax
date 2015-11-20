package config

import (
	"path/filepath"
)

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stores/bytes_int"
	"github.com/timtadh/sfp/stores/bytes_subgraph"
	"github.com/timtadh/sfp/stores/int_int"
	"github.com/timtadh/sfp/stores/int_json"
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

func (c *Config) BytesSubgraphMultiMap(
	name string,
	deserializeValue func([]byte) *goiso.SubGraph,
) (bytes_subgraph.MultiMap, error) {
	if c.Cache == "" {
		return bytes_subgraph.AnonBpTree(bytes_subgraph.Identity, bytes_subgraph.SerializeSubGraph, bytes_subgraph.Identity, deserializeValue)
	} else {
		return bytes_subgraph.NewBpTree(c.CacheFile(name + ".bptree"), bytes_subgraph.Identity, bytes_subgraph.SerializeSubGraph, bytes_subgraph.Identity, deserializeValue)
	}
}

func (c *Config) BytesIntMultiMap(name string) (bytes_int.MultiMap, error) {
	if c.Cache == "" {
		return bytes_int.AnonBpTree()
	} else {
		return bytes_int.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

func (c *Config) IntIntMultiMap(name string) (int_int.MultiMap, error) {
	if c.Cache == "" {
		return int_int.AnonBpTree()
	} else {
		return int_int.NewBpTree(c.CacheFile(name + ".bptree"))
	}
}

func (c *Config) IntJsonMultiMap(name string) (int_json.MultiMap, error) {
	if c.Cache == "" {
		return int_json.AnonBpTree()
	} else {
		return int_json.NewBpTree(c.CacheFile(name + ".bptree"))
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
