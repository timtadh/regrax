package itemsets

import (
	"sync"
)

import (
	"github.com/timtadh/fs2/bptree"
	"github.com/timtadh/fs2/fmap"
)

import (
	
)


type MultiMap interface {
	Keys() (KeyIterator, error)
	Values() (ValueIterator, error)
	Iterate() (Iterator, error)
	Find(key *ItemSet) (Iterator, error)
	Has(key *ItemSet) (bool, error)
	Count(key *ItemSet) (int, error)
	Add(key *ItemSet, value *ItemSet) error
	Remove(key *ItemSet, where func(*ItemSet) bool) error
	Size() int
	Close() error
	Delete() error
}

type Iterator func() (*ItemSet, *ItemSet, error, Iterator)
type KeyIterator func() (*ItemSet, error, KeyIterator)
type ValueIterator func() (*ItemSet, error, ValueIterator)

func Do(run func() (Iterator, error), do func(key *ItemSet, value *ItemSet) error) error {
	kvi, err := run()
	if err != nil {
		return err
	}
	var key *ItemSet
	var value *ItemSet
	for key, value, err, kvi = kvi(); kvi != nil; key, value, err, kvi = kvi() {
		e := do(key, value)
		if e != nil {
			return e
		}
	}
	return err
}

func DoKey(run func() (KeyIterator, error), do func(*ItemSet) error) error {
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

func DoValue(run func() (ValueIterator, error), do func(*ItemSet) error) error {
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


type BpTree struct {
	bf *fmap.BlockFile
	bpt *bptree.BpTree
	mutex sync.Mutex
}

func AnonBpTree() (*BpTree, error) {
	bf, err := fmap.Anonymous(fmap.BLOCKSIZE)
	if err != nil {
		return nil, err
	}
	return newBpTree(bf)
}

func NewBpTree(path string) (*BpTree, error) {
	bf, err := fmap.CreateBlockFile(path)
	if err != nil {
		return nil, err
	}
	return newBpTree(bf)
}

func OpenBpTree(path string) (*BpTree, error) {
	bf, err := fmap.OpenBlockFile(path)
	if err != nil {
		return nil, err
	}
	bpt, err := bptree.Open(bf)
	if err != nil {
		return nil, err
	}
	b := &BpTree{
		bf: bf,
		bpt: bpt,
	}
	return b, nil
}

func newBpTree(bf *fmap.BlockFile) (*BpTree, error) {
	bpt, err := bptree.New(bf, -1, -1)
	if err != nil {
		return nil, err
	}
	b := &BpTree{
		bf: bf,
		bpt: bpt,
	}
	return b, nil
}

func (b *BpTree) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bf.Close()
}

func (b *BpTree) Delete() error {
	err := b.Close()
	if err != nil {
		return err
	}
	if b.bf.Path() != "" {
		return b.bf.Remove()
	}
	return nil
}

func (b *BpTree) Size() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Size()
}

func (b *BpTree) Add(key *ItemSet, val *ItemSet) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Add(ItemSetSerialize(key), ItemSetSerialize(val))
}

func (b *BpTree) Count(key *ItemSet) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Count(ItemSetSerialize(key))
}

func (b *BpTree) Has(key *ItemSet) (bool, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Has(ItemSetSerialize(key))
}

func (b *BpTree) kvIter(kvi bptree.KVIterator) (it Iterator) {
	it = func() (key *ItemSet, value *ItemSet, err error, _ Iterator) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		var k, v []byte
		k, v, err, kvi = kvi()
		if err != nil {
			return nil, nil, err, nil
		}
		if kvi == nil {
			return nil, nil, nil, nil
		}
		key = ItemSetDeserialize(k)
		value = ItemSetDeserialize(v)
		return key, value, nil, it
	}
	return it
}

func (b *BpTree) keyIter(raw bptree.Iterator) (it KeyIterator) {
	it = func() (key *ItemSet, err error, _ KeyIterator) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		var i []byte
		i, err, raw = raw()
		if err != nil {
			return nil, err, nil
		}
		if raw == nil {
			return nil, nil, nil
		}
		key = ItemSetDeserialize(i)
		return key, nil, it
	}
	return it
}

func (b *BpTree) valueIter(raw bptree.Iterator) (it ValueIterator) {
	it = func() (value *ItemSet, err error, _ ValueIterator) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		var i []byte
		i, err, raw = raw()
		if err != nil {
			return nil, err, nil
		}
		if raw == nil {
			return nil, nil, nil
		}
		value = ItemSetDeserialize(i)
		return value, nil, it
	}
	return it
}

func (b *BpTree) Keys() (it KeyIterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Keys()
	if err != nil {
		return nil, err
	}
	return b.keyIter(raw), nil
}

func (b *BpTree) Values() (it ValueIterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Values()
	if err != nil {
		return nil, err
	}
	return b.valueIter(raw), nil
}

func (b *BpTree) Find(key *ItemSet) (it Iterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Find(ItemSetSerialize(key))
	if err != nil {
		return nil, err
	}
	return b.kvIter(raw), nil
}

func (b *BpTree) Iterate() (it Iterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Iterate()
	if err != nil {
		return nil, err
	}
	return b.kvIter(raw), nil
}

func (b *BpTree) Remove(key *ItemSet, where func(*ItemSet) bool) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Remove(ItemSetSerialize(key), func(bytes []byte) bool {
		return where(ItemSetDeserialize(bytes))
	})
}

