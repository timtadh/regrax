//go:generate fs2-generic --output=wrapper.go --package-name=itemsets bptree --key-type=*ItemSet --key-serializer=ItemSetSerialize --key-deserializer=ItemSetDeserialize --value-type=*ItemSet --value-serializer=ItemSetSerialize --value-deserializer=ItemSetDeserialize
package itemsets

import (
	"encoding/binary"
)


type ItemSet struct {
	Items []int32
	Txs []int32
}

func ItemSetSerialize(i *ItemSet) []byte {
	bytes := make([]byte, 4*(len(i.Items) + len(i.Txs) + 2))
	binary.BigEndian.PutUint32(bytes[0:4], uint32(len(i.Items)))
	binary.BigEndian.PutUint32(bytes[4:8], uint32(len(i.Txs)))
	s := 8
	e := s + 4
	for _, item := range i.Items {
		binary.BigEndian.PutUint32(bytes[s:e], uint32(item))
		s += 4
		e = s + 4
	}
	for _, tx := range i.Txs {
		binary.BigEndian.PutUint32(bytes[s:e], uint32(tx))
		s += 4
		e = s + 4
	}
	return bytes
}

func ItemSetDeserialize(bytes []byte) *ItemSet {
	lenItems := int(binary.BigEndian.Uint32(bytes[0:4]))
	lenTxs := int(binary.BigEndian.Uint32(bytes[4:8]))
	s := 8
	e := s + 4
	i := &ItemSet{
		Items: make([]int32, 0, lenItems),
		Txs: make([]int32, 0, lenTxs),
	}
	for x := 0; x < lenItems; x++ {
		item := int32(binary.BigEndian.Uint32(bytes[s:e]))
		i.Items = append(i.Items, item)
		s += 4
		e = s + 4
	}
	for x := 0; x < lenTxs; x++ {
		tx := int32(binary.BigEndian.Uint32(bytes[s:e]))
		i.Txs = append(i.Txs, tx)
		s += 4
		e = s + 4
	}
	return i
}

