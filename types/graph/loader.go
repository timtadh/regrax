package graph

import (
)

import (
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
)


type MakeLoader func(*Graph) lattice.Loader

type Graph struct {
	makeLoader MakeLoader
	config *config.Config
}

func NewGraph(config *config.Config, makeLoader MakeLoader) (g *Graph, err error) {
	g = &Graph{
		makeLoader: makeLoader,
		config: config,
	}
	return g, nil
}

func (g *Graph) Loader() lattice.Loader {
	return g.makeLoader(g)
}

func (g *Graph) Close() error {
	return nil
}

func NewVegLoader(g *Graph) lattice.Loader {
	return nil
}
