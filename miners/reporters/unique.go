package reporters

import (
	"fmt"
	"io"
	"os"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/stores/bytes_int"
)

type Unique struct {
	count     int
	fmtr      lattice.Formatter
	Seen      bytes_int.MultiMap
	Reporter  miners.Reporter
	histogram io.WriteCloser
}

func NewUnique(conf *config.Config, fmtr lattice.Formatter, reporter miners.Reporter, histogramName string) (*Unique, error) {
	seen, err := conf.BytesIntMultiMap("unique-seen")
	if err != nil {
		return nil, err
	}
	var histogram io.WriteCloser = nil
	if histogramName != "" {
		histogram, err = os.Create(conf.OutputFile(histogramName + ".csv"))
		if err != nil {
			return nil, err
		}
	}
	u := &Unique{
		fmtr:      fmtr,
		Seen:      seen,
		Reporter:  reporter,
		histogram: histogram,
	}
	return u, nil
}

func (r *Unique) Report(n lattice.Node) error {
	r.count++
	label := []byte(r.fmtr.PatternName(n))
	if has, err := r.Seen.Has(label); err != nil {
		return err
	} else if has {
		var count int32
		err = r.Seen.DoFind(label, func(_ []byte, c int32) error {
			count = c
			return nil
		})
		if err != nil {
			return err
		}
		err = r.Seen.Remove(label, func(_ int32) bool { return true })
		if err != nil {
			return err
		}
		return r.Seen.Add(label, count+1)
	} else {
		err = r.Seen.Add(label, 1)
		if err != nil {
			return nil
		}
		return r.Reporter.Report(n)
	}
}

func (r *Unique) Close() error {
	if r.histogram != nil {
		err := bytes_int.Do(r.Seen.Iterate, func(k []byte, c int32) error {
			name := string(k)
			fmt.Fprintf(r.histogram, "%d, %.5g, %v\n", c, float64(c)/float64(r.count), name)
			return nil
		})
		if err != nil {
			errors.Logf("ERROR", "%v", err)
		}
		err = r.histogram.Close()
		if err != nil {
			errors.Logf("ERROR", "%v", err)
		}
	}
	err := r.Seen.Delete()
	if err != nil {
		errors.Logf("ERROR", "%v", err)
	}
	return r.Reporter.Close()
}
