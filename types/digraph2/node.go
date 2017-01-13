package digraph2

import (
	"github.com/timtadh/sfp/types/digraph2/subgraph"
)

type Node struct {
	Pattern
}

type Pattern struct {
	SubGraph subgraph.SubGraph
}
