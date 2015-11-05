package itemsets

import (
	"encoding/binary"
)


type ItemSet struct {
	Items []int32
	Txs []int32
}

func (i *ItemSet) Serialize() []byte {
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

func ItemSetFromBytes(bytes []byte) *ItemSet {
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

type MultiMap interface {
	Keys() (ItemSetIterator, error)
	Values() (ItemSetIterator, error)
	Iterate() (Iterator, error)
	Find(key *ItemSet) (Iterator, error)
	Has(key *ItemSet) (bool, error)
	Count(key *ItemSet) (int, error)
	Add(key, value *ItemSet) error
	Remove(key *ItemSet, where func(*ItemSet) bool) error
	Size() int
	Close() error
	Delete() error
}

type Iterator func() (*ItemSet, *ItemSet, error, Iterator)
type ItemSetIterator func() (*ItemSet, error, ItemSetIterator)

func Do(run func() (Iterator, error), do func(key, value *ItemSet) error) error {
	kvi, err := run()
	if err != nil {
		return err
	}
	var key, value *ItemSet
	for key, value, err, kvi = kvi(); kvi != nil; key, value, err, kvi = kvi() {
		e := do(key, value)
		if e != nil {
			return e
		}
	}
	return err
}

func DoItemSets(run func() (ItemSetIterator, error), do func(*ItemSet) error) error {
	it, err := run()
	if err != nil {
		return err
	}
	var item *ItemSet
	for item, err, it = it(); it != nil; item, err, it = it() {
		e := do(item)
		if e != nil {
			return e
		}
	}
	return err
}

