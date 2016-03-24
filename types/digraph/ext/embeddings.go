package ext

import (
	"bytes"
	"sort"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/goiso"
)

type Embeddings []*goiso.SubGraph

func (sgs Embeddings) Len() int {
	return len(sgs)
}

func (sgs Embeddings) Less(i, j int) bool {
	return bytes.Compare(sgs[i].ShortLabel(), sgs[j].ShortLabel()) < 0
}

func (sgs Embeddings) Swap(i, j int) {
	sgs[i], sgs[j] = sgs[j], sgs[i]
}

func (sgs Embeddings) Verify() error {
	if len(sgs) <= 0 {
		return errors.Errorf("empty partition")
	}
	label := sgs[0].ShortLabel()
	for _, sg := range sgs {
		if !bytes.Equal(label, sg.ShortLabel()) {
			return errors.Errorf("bad partition %v %v", sgs[0].Label(), sg.Label())
		}
	}
	return nil
}

func (sgs Embeddings) Partition() []Embeddings {
	sort.Sort(sgs)
	parts := make([]Embeddings, 0, 10)
	add := func(parts []Embeddings, buf Embeddings) []Embeddings {
		err := buf.Verify()
		if err != nil {
			errors.Logf("ERROR", "%v", err)
		} else {
			parts = append(parts, buf)
		}
		return parts
	}
	buf := make(Embeddings, 0, 10)
	var ckey []byte = nil
	for _, sg := range sgs {
		label := sg.ShortLabel()
		if ckey != nil && !bytes.Equal(ckey, label) {
			parts = add(parts, buf)
			buf = make(Embeddings, 0, 10)
		}
		ckey = label
		buf = append(buf, sg)
	}
	if len(buf) > 0 {
		parts = add(parts, buf)
	}
	return parts
}
