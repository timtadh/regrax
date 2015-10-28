package intint


type MultiMap interface {
	Keys() (IntIterator, error)
	Values() (IntIterator, error)
	Iterate() (Iterator, error)
	Find(key int) (Iterator, error)
	Has(key int) (bool, error)
	Count(key int) (int, error)
	Add(key, value int) error
	Remove(key int, where func(int) bool) error
	Size() int
	Close() error
	Delete() error
}

type Iterator func() (int, int, error, Iterator)
type IntIterator func() (int, error, IntIterator)

func Do(run func() (Iterator, error), do func(key, value int) error) error {
	kvi, err := run()
	if err != nil {
		return err
	}
	var key, value int
	for key, value, err, kvi = kvi(); kvi != nil; key, value, err, kvi = kvi() {
		e := do(key, value)
		if e != nil {
			return e
		}
	}
	return err
}

func DoInt(run func() (IntIterator, error), do func(int) error) error {
	it, err := run()
	if err != nil {
		return err
	}
	var item int
	for item, err, it = it(); it != nil; item, err, it = it() {
		e := do(item)
		if e != nil {
			return e
		}
	}
	return err
}

