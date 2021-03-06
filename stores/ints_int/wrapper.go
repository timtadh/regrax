package ints_int

/*
* This file is machine generated by fs2-generic. You can obtain
* fs2-generic from github.com/timtadh/fs2/fs2-generic . To install
*
*   $ go get github.com/timtadh/fs2
*   $ go install github.com/timtadh/fs2/fs2-generic
*
* The command used to generate the file was
*
*   $ fs2-generic \
*     --output=wrapper.go \
*     --package-name=ints_int \
*     bptree \
*     --key-type=[]int32 \
*     --key-serializer=github.com/timtadh/regrax/stores/ints_ints/SerializeInt32s \
*     --key-deserializer=github.com/timtadh/regrax/stores/ints_ints/DeserializeInt32s \
*     --value-type=int32 \
*     --value-empty=0 \
*     --value-size=4 \
*     --value-serializer=github.com/timtadh/regrax/stores/int_int/SerializeInt32 \
*     --value-deserializer=github.com/timtadh/regrax/stores/int_int/DeserializeInt32
*
* By including this code (and linking to github.com/timtadh/fs2) you
* accept the terms of the GPL version 3 unless another license has been
* obtained in writing from the author(s) of the package. The list of
* authors and full licensing information is located at
* https://github.com/timtadh/fs2/LICENSE
*
* This library is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 3 of the License, or (at
* your option) any later version.
*
* This library is distributed in the hope that it will be useful, but
* WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
* General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this library; if not, write to the Free Software
* Foundation, Inc.,
*   51 Franklin Street, Fifth Floor,
*   Boston, MA  02110-1301
*   USA
 */

import (
	"sync"
)

import (
	"github.com/timtadh/fs2"
	"github.com/timtadh/fs2/bptree"
	"github.com/timtadh/fs2/fmap"
)

import (
	"github.com/timtadh/regrax/stores/int_int"
	"github.com/timtadh/regrax/stores/ints_ints"
)

type MultiMap interface {
	Keys() (KeyIterator, error)
	Values() (ValueIterator, error)
	Iterate() (Iterator, error)
	Backward() (Iterator, error)
	Find(key []int32) (Iterator, error)
	DoFind(key []int32, do func([]int32, int32) error) error
	Range(from, to []int32) (Iterator, error)
	DoRange(from, to []int32, do func([]int32, int32) error) error
	Has(key []int32) (bool, error)
	Count(key []int32) (int, error)
	Add(key []int32, value int32) error
	Remove(key []int32, where func(int32) bool) error
	Size() int
	Close() error
	Delete() error
}

type Iterator func() ([]int32, int32, error, Iterator)
type KeyIterator func() ([]int32, error, KeyIterator)
type ValueIterator func() (int32, error, ValueIterator)

func Do(run func() (Iterator, error), do func(key []int32, value int32) error) error {
	kvi, err := run()
	if err != nil {
		return err
	}
	var key []int32
	var value int32
	for key, value, err, kvi = kvi(); kvi != nil; key, value, err, kvi = kvi() {
		e := do(key, value)
		if e != nil {
			return e
		}
	}
	return err
}

func DoKey(run func() (KeyIterator, error), do func([]int32) error) error {
	it, err := run()
	if err != nil {
		return err
	}
	var item []int32
	for item, err, it = it(); it != nil; item, err, it = it() {
		e := do(item)
		if e != nil {
			return e
		}
	}
	return err
}

func DoValue(run func() (ValueIterator, error), do func(int32) error) error {
	it, err := run()
	if err != nil {
		return err
	}
	var item int32
	for item, err, it = it(); it != nil; item, err, it = it() {
		e := do(item)
		if e != nil {
			return e
		}
	}
	return err
}

type BpTree struct {
	bf    *fmap.BlockFile
	bpt   *bptree.BpTree
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
		bf:  bf,
		bpt: bpt,
	}
	return b, nil
}

func newBpTree(bf *fmap.BlockFile) (*BpTree, error) {
	bpt, err := bptree.New(bf, -1, 4)
	if err != nil {
		return nil, err
	}
	b := &BpTree{
		bf:  bf,
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

func (b *BpTree) Add(key []int32, val int32) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Add(ints_ints.SerializeInt32s(key), int_int.SerializeInt32(val))
}

func (b *BpTree) Count(key []int32) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Count(ints_ints.SerializeInt32s(key))
}

func (b *BpTree) Has(key []int32) (bool, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Has(ints_ints.SerializeInt32s(key))
}

func (b *BpTree) kvIter(kvi fs2.Iterator) (it Iterator) {
	it = func() (key []int32, value int32, err error, _ Iterator) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		var k, v []byte
		k, v, err, kvi = kvi()
		if err != nil {
			return nil, 0, err, nil
		}
		if kvi == nil {
			return nil, 0, nil, nil
		}
		key = ints_ints.DeserializeInt32s(k)
		value = int_int.DeserializeInt32(v)
		return key, value, nil, it
	}
	return it
}

func (b *BpTree) keyIter(raw fs2.ItemIterator) (it KeyIterator) {
	it = func() (key []int32, err error, _ KeyIterator) {
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
		key = ints_ints.DeserializeInt32s(i)
		return key, nil, it
	}
	return it
}

func (b *BpTree) valueIter(raw fs2.ItemIterator) (it ValueIterator) {
	it = func() (value int32, err error, _ ValueIterator) {
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
		value = int_int.DeserializeInt32(i)
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

func (b *BpTree) Find(key []int32) (it Iterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Find(ints_ints.SerializeInt32s(key))
	if err != nil {
		return nil, err
	}
	return b.kvIter(raw), nil
}

func (b *BpTree) DoFind(key []int32, do func([]int32, int32) error) error {
	return Do(func() (Iterator, error) { return b.Find(key) }, do)
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

func (b *BpTree) Range(from, to []int32) (it Iterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Range(ints_ints.SerializeInt32s(from), ints_ints.SerializeInt32s(to))
	if err != nil {
		return nil, err
	}
	return b.kvIter(raw), nil
}

func (b *BpTree) DoRange(from, to []int32, do func([]int32, int32) error) error {
	return Do(func() (Iterator, error) { return b.Range(from, to) }, do)
}

func (b *BpTree) Backward() (it Iterator, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	raw, err := b.bpt.Backward()
	if err != nil {
		return nil, err
	}
	return b.kvIter(raw), nil
}

func (b *BpTree) Remove(key []int32, where func(int32) bool) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bpt.Remove(ints_ints.SerializeInt32s(key), func(bytes []byte) bool {
		return where(int_int.DeserializeInt32(bytes))
	})
}
