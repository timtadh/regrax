//go:generate fs2-generic --output=wrapper.go --package-name=ints_ints bptree --key-type=[]int32 --key-serializer=SerializeInt32s --key-deserializer=DeserializeInt32s --value-type=[]int32 --value-serializer=SerializeInt32s --value-deserializer=DeserializeInt32s
package ints_ints

import (
	"encoding/binary"
)

func SerializeInt32s(list []int32) []byte {
	bytes := make([]byte, 4*(1 + len(list)))
	binary.BigEndian.PutUint32(bytes[0:4], uint32(len(list)))
	s := 4
	e := s + 4
	for _, i := range list {
		binary.BigEndian.PutUint32(bytes[s:e], uint32(i))
		s += 4
		e = s + 4
	}
	return bytes
}

func DeserializeInt32s(bytes []byte) []int32 {
	lenList := int(binary.BigEndian.Uint32(bytes[0:4]))
	list := make([]int32, 0, lenList)
	s := 4
	e := s + 4
	for x := 0; x < lenList; x++ {
		item := int32(binary.BigEndian.Uint32(bytes[s:e]))
		list = append(list, item)
		s += 4
		e = s + 4
	}
	return list
}

