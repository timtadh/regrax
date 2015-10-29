package intint


type MultiMap interface {
	Keys() (IntIterator, error)
	Values() (IntIterator, error)
	Iterate() (Iterator, error)
	Find(key int32) (Iterator, error)
	Has(key int32) (bool, error)
	Count(key int32) (int, error)
	Add(key, value int32) error
	Remove(key int32, where func(int32) bool) error
	Size() int
	Close() error
	Delete() error
}

type Iterator func() (int32, int32, error, Iterator)
type IntIterator func() (int32, error, IntIterator)

func Do(run func() (Iterator, error), do func(key, value int32) error) error {
	kvi, err := run()
	if err != nil {
		return err
	}
	var key, value int32
	for key, value, err, kvi = kvi(); kvi != nil; key, value, err, kvi = kvi() {
		e := do(key, value)
		if e != nil {
			return e
		}
	}
	return err
}

func DoInt(run func() (IntIterator, error), do func(int32) error) error {
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

