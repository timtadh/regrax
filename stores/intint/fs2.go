package intint

import (
	"sync"
)

import(
	"github.com/timtadh/fs2/bptree"
	"github.com/timtadh/fs2/fmap"
	"github.com/timtadh/fs2/slice"
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
	bpt, err := bptree.New(bf, 8, 8)
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

func (b *BpTree) Add(key, value int) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k := int64(key)
	v := int64(value)
	return b.bpt.Add(slice.Int64AsSlice(&k), slice.Int64AsSlice(&v))
}

func (b *BpTree) Count(key int) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k := int64(key)
	return b.bpt.Count(slice.Int64AsSlice(&k))
}

func (b *BpTree) Has(key int) (bool, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k := int64(key)
	return b.bpt.Has(slice.Int64AsSlice(&k))
}

func (b *BpTree) kvIter(kvi bptree.KVIterator) (it Iterator) {
	it = func() (key int, value int, err error, _ Iterator) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		var k, v []byte
		k, v, err, kvi = kvi()
		if err != nil {
			return 0, 0, err, nil
		}
		if kvi == nil {
			return 0, 0, nil, nil
		}
		key = int(*slice.AsInt64(&k))
		value = int(*slice.AsInt64(&v))
		return key, value, nil, it
	}
	return it
}

func (b *BpTree) itemIter(raw bptree.Iterator) (it IntIterator) {
	it = func() (item int, err error, _ IntIterator) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		var i []byte
		i, err, raw = raw()
		if err != nil {
			return 0, err, nil
		}
		if raw == nil {
			return 0, nil, nil
		}
		item = int(*slice.AsInt64(&i))
		return item, nil, it
	}
	return it
}

func (b *BpTree) Keys() (it IntIterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Keys()
	if err != nil {
		return nil, err
	}
	return b.itemIter(raw), nil
}

func (b *BpTree) Values() (it IntIterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Values()
	if err != nil {
		return nil, err
	}
	return b.itemIter(raw), nil
}

func (b *BpTree) Find(key int) (it Iterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k := int64(key)
	raw, err := b.bpt.Find(slice.Int64AsSlice(&k))
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

func (b *BpTree) Remove(key int, where func(int) bool) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k := int64(key)
	err := b.bpt.Remove(slice.Int64AsSlice(&k), func(bytes []byte) bool {
		return where(int(*slice.AsInt64(&bytes)))
	})
	if err == nil {
		return err
	}
	return nil
}
