//go:generate fs2-generic --output=wrapper.go --package-name=bytes_extension bptree --key-type=[]byte --key-serializer=github.com/timtadh/regrax/stores/bytes_subgraph/Identity --key-deserializer=github.com/timtadh/regrax/stores/bytes_subgraph/Identity --value-type=*github.com/timtadh/regrax/types/digraph/subgraph/Extension --value-size=20 --value-serializer=SerializeExtension --value-deserializer=DeserializeExtension
package bytes_extension

import (
	"encoding/binary"
)

import (
	"github.com/timtadh/regrax/types/digraph/subgraph"
)

func SerializeExtension(e *subgraph.Extension) []byte {
	bytes := make([]byte, 20)
	binary.BigEndian.PutUint32(bytes[0:4], uint32(e.Source.Idx))
	binary.BigEndian.PutUint32(bytes[4:8], uint32(e.Source.Color))
	binary.BigEndian.PutUint32(bytes[8:12], uint32(e.Target.Idx))
	binary.BigEndian.PutUint32(bytes[12:16], uint32(e.Target.Color))
	binary.BigEndian.PutUint32(bytes[16:20], uint32(e.Color))
	return bytes
}

func DeserializeExtension(bytes []byte) *subgraph.Extension {
	srcIdx := int(binary.BigEndian.Uint32(bytes[0:4]))
	srcColor := int(binary.BigEndian.Uint32(bytes[4:8]))
	targIdx := int(binary.BigEndian.Uint32(bytes[8:12]))
	targColor := int(binary.BigEndian.Uint32(bytes[12:16]))
	color := int(binary.BigEndian.Uint32(bytes[16:20]))
	return subgraph.NewExt(
		subgraph.Vertex{Idx: srcIdx, Color: srcColor},
		subgraph.Vertex{Idx: targIdx, Color: targColor},
		color,
	)
}
