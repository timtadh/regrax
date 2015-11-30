//go:generate fs2-generic --output=wrapper.go --package-name=bytes_subgraph bptree --use-parameterized-serialization --key-type=[]byte --value-type=*github.com/timtadh/goiso/SubGraph
package bytes_subgraph

import (
	"github.com/timtadh/goiso"
)

func Identity(in []byte) []byte { return in }

func DeserializeSubGraph(g *goiso.Graph) func([]byte) *goiso.SubGraph {
	return func(bytes []byte) *goiso.SubGraph {
		return goiso.DeserializeSubGraph(g, bytes)
	}
}

func SerializeSubGraph(sg *goiso.SubGraph) []byte {
	return sg.Serialize()
}
