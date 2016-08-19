package config

import (
	"math/rand"
	"path/filepath"
	"sync"
	"runtime"
)

import (
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/stacks/subgraph"
	"github.com/timtadh/sfp/stores/bytes_bytes"
	"github.com/timtadh/sfp/stores/bytes_extension"
	"github.com/timtadh/sfp/stores/bytes_float"
	"github.com/timtadh/sfp/stores/bytes_int"
	"github.com/timtadh/sfp/stores/bytes_subgraph"
	"github.com/timtadh/sfp/stores/int_int"
	"github.com/timtadh/sfp/stores/int_json"
	"github.com/timtadh/sfp/stores/ints_int"
	"github.com/timtadh/sfp/stores/ints_ints"
	"github.com/timtadh/sfp/stores/subgraph_embedding"
	"github.com/timtadh/sfp/stores/subgraph_overlap"
)

type Config struct {
	Cache            string
	Output           string
	Support, Samples int
	Unique           bool
	Parallelism      int
	AsyncTasks       sync.WaitGroup
}

func (c *Config) Copy() *Config {
	return &Config{
		Cache:   c.Cache,
		Output:  c.Output,
		Support: c.Support,
		Samples: c.Samples,
		Unique:  c.Unique,
	}
}

func (c *Config) Workers() int {
	if c.Parallelism == 0 {
		return 1
	} else if c.Parallelism == -1 {
		return runtime.NumCPU()
	} else {
		return c.Parallelism
	}

}

func (c *Config) Randstr() string {
	runes := make([]rune, 0, 10)
	for i := 0; i < 10; i++ {
		runes = append(runes, rune(97+rand.Intn(26)))
	}
	return string(runes)
}

func (c *Config) CacheFile(name string) string {
	return filepath.Join(c.Cache, name)
}

func (c *Config) OutputFile(name string) string {
	return filepath.Join(c.Output, name)
}

func (c *Config) MultiMap(name string) (bytes_bytes.MultiMap, error) {
	if c.Cache == "" {
		return bytes_bytes.AnonBpTree()
	} else {
		return bytes_bytes.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) SubgraphList(
	name string,
	deserializeValue func([]byte) *goiso.SubGraph,
) (subgraph.List, error) {
	if c.Cache == "" {
		return subgraph.AnonList(bytes_subgraph.SerializeSubGraph, deserializeValue)
	} else {
		return subgraph.NewList(c.CacheFile(name+"-"+c.Randstr()+".mmlist"), bytes_subgraph.SerializeSubGraph, deserializeValue)
	}
}

func (c *Config) BytesSubgraphMultiMap(
	name string,
	deserializeValue func([]byte) *goiso.SubGraph,
) (bytes_subgraph.MultiMap, error) {
	if c.Cache == "" {
		return bytes_subgraph.AnonBpTree(bytes_subgraph.Identity, bytes_subgraph.SerializeSubGraph, bytes_subgraph.Identity, deserializeValue)
	} else {
		return bytes_subgraph.NewBpTree(c.CacheFile(name+"-"+c.Randstr()+".bptree"), bytes_subgraph.Identity, bytes_subgraph.SerializeSubGraph, bytes_subgraph.Identity, deserializeValue)
	}
}

func (c *Config) BytesExtensionMultiMap(name string) (bytes_extension.MultiMap, error) {
	if c.Cache == "" {
		return bytes_extension.AnonBpTree()
	} else {
		return bytes_extension.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) BytesFloatMultiMap(name string) (bytes_float.MultiMap, error) {
	if c.Cache == "" {
		return bytes_float.AnonBpTree()
	} else {
		return bytes_float.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) BytesIntMultiMap(name string) (bytes_int.MultiMap, error) {
	if c.Cache == "" {
		return bytes_int.AnonBpTree()
	} else {
		return bytes_int.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) IntIntMultiMap(name string) (int_int.MultiMap, error) {
	if c.Cache == "" {
		return int_int.AnonBpTree()
	} else {
		return int_int.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) IntJsonMultiMap(name string) (int_json.MultiMap, error) {
	if c.Cache == "" {
		return int_json.AnonBpTree()
	} else {
		return int_json.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) IntsIntMultiMap(name string) (ints_int.MultiMap, error) {
	if c.Cache == "" {
		return ints_int.AnonBpTree()
	} else {
		return ints_int.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) IntsIntsMultiMap(name string) (ints_ints.MultiMap, error) {
	if c.Cache == "" {
		return ints_ints.AnonBpTree()
	} else {
		return ints_ints.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) SubgraphEmbeddingMultiMap(name string) (subgraph_embedding.MultiMap, error) {
	if c.Cache == "" {
		return subgraph_embedding.AnonBpTree()
	} else {
		return subgraph_embedding.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}

func (c *Config) SubgraphOverlapMultiMap(name string) (subgraph_overlap.MultiMap, error) {
	if c.Cache == "" {
		return subgraph_overlap.AnonBpTree()
	} else {
		return subgraph_overlap.NewBpTree(c.CacheFile(name + "-" + c.Randstr() + ".bptree"))
	}
}
