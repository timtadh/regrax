package digraph

import (
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph/ext"
	"github.com/timtadh/sfp/types/digraph/support"
)

// YOU ARE HERE:
//
// Ok so what we need to do is take an a pattern (eg. subgraph.SubGraph) and
// compute all of the unique subgraph.Extension(s). Those will then get stored
// in digraph.Extension.
//
// Then when computing children the extensions are looked up rather than found
// using the embeddings list. Furthermore, the only supported embeddings are
// stored in the embedding list to cut down on space.
//
// kk.



func nodesFromEmbeddings(n Node, embs ext.Embeddings) (nodes []lattice.Node, err error) {
	dt := n.dt()
	partitioned := embs.Partition()
	sum := 0
	for _, sgs := range partitioned {
		sum += len(sgs)
		new_node := n.New(nil, support.Dedup(sgs))
		if len(sgs) < dt.Support() {
			continue
		}
		new_embeddings, err := new_node.Embeddings()
		if err != nil {
			return nil, err
		}
		supported, err := dt.Supported(dt, new_embeddings)
		if err != nil {
			return nil, err
		}
		if len(supported) >= dt.Support() {
			nodes = append(nodes, new_node)
		}
	}
	// errors.Logf("DEBUG", "sum(len(partition)) %v", sum)
	// errors.Logf("DEBUG", "kids of %v are %v", n, nodes)
	return nodes, cache(dt, dt.ChildCount, dt.Children, n.Label(), nodes)
}

