//go:generate fs2-generic --output=wrapper.go --package-name=bytes_float bptree --key-type=[]byte --key-serializer=github.com/timtadh/sfp/stores/bytes_subgraph/Identity --key-deserializer=github.com/timtadh/sfp/stores/bytes_subgraph/Identity --value-type=float64 --value-empty=0.0 --value-size=8 --value-serializer=SerializeFloat64 --value-deserializer=DeserializeFloat64
package bytes_float

import (
	"encoding/binary"
	"math"
)

func SerializeFloat64(f float64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, math.Float64bits(f))
	return bytes
}

func DeserializeFloat64(bytes []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(bytes))
}
