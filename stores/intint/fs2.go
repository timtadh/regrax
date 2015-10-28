package intint

import (
	"sync"
)

import(
	"github.com/timtadh/data-structures/types"
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

var marshal types.ItemMarshal
var unmarshal types.ItemUnmarshal

func init() {
	marshal, unmarshal = types.Int64Marshals()
}

func (b *BpTree) Add(key, val int) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k, err := marshal(types.Int64(int64(key)))
	if err != nil {
		return err
	}
	v, err := marshal(types.Int64(int64(val)))
	if err != nil {
		return err
	}
	return b.bpt.Add(k, v)
}

func (b *BpTree) Count(key int) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k, err := marshal(types.Int64(int64(key)))
	if err != nil {
		return 0, err
	}
	return b.bpt.Count(k)
}

func (b *BpTree) Has(key int) (bool, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	k, err := marshal(types.Int64(int64(key)))
	if err != nil {
		return false, err
	}
	return b.bpt.Has(k)
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
		K, err := unmarshal(k)
		if err != nil {
			return 0, 0, err, nil
		}
		V, err := unmarshal(v)
		if err != nil {
			return 0, 0, err, nil
		}
		key = int(int64(K.(types.Int64)))
		value = int(int64(V.(types.Int64)))
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
		I, err := unmarshal(i)
		if err != nil {
			return 0, err, nil
		}
		item = int(int64(I.(types.Int64)))
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
	k, err := marshal(types.Int64(int64(key)))
	if err != nil {
		return nil, err
	}
	raw, err := b.bpt.Find(k)
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
	k, err := marshal(types.Int64(int64(key)))
	if err != nil {
		return err
	}
	rerr := b.bpt.Remove(k, func(bytes []byte) bool {
		I, e := unmarshal(bytes)
		if e != nil && err == nil {
			err = e
		}
		if e != nil {
			return false
		}
		item := int(int64(I.(types.Int64)))
		return where(item)
	})
	if err != nil {
		return err
	}
	if rerr == nil {
		return rerr
	}
	return nil
}

