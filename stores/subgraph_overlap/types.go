//go:generate fs2-generic --output=wrapper.go --package-name=subgraph_overlap bptree --key-type=*github.com/timtadh/sfp/types/digraph/subgraph/SubGraph --key-serializer=github.com/timtadh/sfp/stores/subgraph_embedding/SerializeSubGraph --key-deserializer=github.com/timtadh/sfp/stores/subgraph_embedding/DeserializeSubGraph --value-type=[]map[int]bool --value-serializer=SerializeOverlap --value-deserializer=DeserializeOverlap

package subgraph_overlap

import (
	"encoding/binary"
)

import ()

func SerializeOverlap(overlap []map[int]bool) []byte {
	size := 4
	for _, o := range overlap {
		size += 4 * (1 + len(o))
	}
	bytes := make([]byte, size)
	binary.BigEndian.PutUint32(bytes[0:4], uint32(len(overlap)))
	s := 4
	e := s + 4
	for _, o := range overlap {
		binary.BigEndian.PutUint32(bytes[s:e], uint32(len(o)))
		s += 4
		e = s + 4
		for i := range o {
			binary.BigEndian.PutUint32(bytes[s:e], uint32(i))
			s += 4
			e = s + 4
		}
	}
	return bytes
}

func DeserializeOverlap(bytes []byte) []map[int]bool {
	lenOverlap := int(binary.BigEndian.Uint32(bytes[0:4]))
	overlap := make([]map[int]bool, 0, lenOverlap)
	s := 4
	e := s + 4
	for i := 0; i < lenOverlap; i++ {
		lenList := int(binary.BigEndian.Uint32(bytes[s:e]))
		o := make(map[int]bool, lenList)
		s += 4
		e = s + 4
		for j := 0; j < lenList; j++ {
			item := int(binary.BigEndian.Uint32(bytes[s:e]))
			o[item] = true
			s += 4
			e = s + 4
		}
		overlap = append(overlap, o)
	}
	return overlap
}
