package itemsets

import (
	"sync"
)

import(
	"github.com/timtadh/fs2/bptree"
	"github.com/timtadh/fs2/fmap"
)


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

func (b *BpTree) Add(key, val *ItemSet) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Add(key.Serialize(), val.Serialize())
}

func (b *BpTree) Count(key *ItemSet) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Count(key.Serialize())
}

func (b *BpTree) Has(key *ItemSet) (bool, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Has(key.Serialize())
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
		key = ItemSetFromBytes(k)
		value = ItemSetFromBytes(v)
		return key, value, nil, it
	}
	return it
}

func (b *BpTree) itemIter(raw bptree.Iterator) (it ItemSetIterator) {
	it = func() (item *ItemSet, err error, _ ItemSetIterator) {
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
		item = ItemSetFromBytes(i)
		return item, nil, it
	}
	return it
}

func (b *BpTree) Keys() (it ItemSetIterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Keys()
	if err != nil {
		return nil, err
	}
	return b.itemIter(raw), nil
}

func (b *BpTree) Values() (it ItemSetIterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Values()
	if err != nil {
		return nil, err
	}
	return b.itemIter(raw), nil
}

func (b *BpTree) Find(key *ItemSet) (it Iterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Find(key.Serialize())
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
	return b.bpt.Remove(key.Serialize(), func(bytes []byte) bool {
		return where(ItemSetFromBytes(bytes))
	})
}

