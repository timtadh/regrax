//go:generate fs2-generic --output=wrapper.go --package-name=int_edge bptree --key-type=int32 --key-size=4 --key-empty=0 --key-serializer=github.com/timtadh/sfp/stores/int_int/SerializeInt32 --key-deserializer=github.com/timtadh/sfp/stores/int_int/DeserializeInt32 --value-type=github.com/timtadh/sfp/types/digraph/subgraph/Edge --value-size=12 --value-empty=subgraph.Edge{} --value-serializer=SerializeEdge --value-deserializer=DeserializeEdge

package int_edge

import (
	"encoding/binary"
)

import (
	"github.com/timtadh/sfp/types/digraph/subgraph"
)


func SerializeEdge(e subgraph.Edge) []byte {
	bytes := make([]byte, 12)
	binary.BigEndian.PutUint32(bytes[0:4], uint32(e.Src))
	binary.BigEndian.PutUint32(bytes[4:8], uint32(e.Color))
	binary.BigEndian.PutUint32(bytes[8:12], uint32(e.Targ))
	return bytes
}

func DeserializeEdge(bytes []byte) subgraph.Edge {
	src := int(binary.BigEndian.Uint32(bytes[0:4]))
	color := int(binary.BigEndian.Uint32(bytes[4:8]))
	targ := int(binary.BigEndian.Uint32(bytes[8:12]))
	return subgraph.Edge{Src:src, Targ:targ, Color:color}
}
