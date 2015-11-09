//go:generate fs2-generic --output=wrapper.go --package-name=intint bptree --key-type=int32 --key-size=4 --key-empty=0 --key-serializer=SerializeInt32 --key-deserializer=DeserializeInt32 --value-type=int32 --value-size=4 --value-empty=0 --value-serializer=SerializeInt32 --value-deserializer=DeserializeInt32
package intint

import (
	"encoding/binary"
)

func SerializeInt32(i int32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(i))
	return bytes
}

func DeserializeInt32(bytes []byte) int32 {
	return int32(binary.BigEndian.Uint32(bytes))
}

