package premusk

import (
	"math/rand"
)

import (
	"github.com/timtadh/data-structures/errors"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/musk"
	"github.com/timtadh/sfp/miners/ospace"
	"github.com/timtadh/sfp/reporters"
	"github.com/timtadh/sfp/miners/walker"
)

type Walker struct {
	walker.Walker
	Teleports           []lattice.Node
	TeleportProbability float64
	teleportAllowed     bool
}

func NewWalker(conf *config.Config, teleportProbability float64) *Walker {
	errors.Logf("INFO", "teleport probability %v", teleportProbability)
	miner := &Walker{
		TeleportProbability: teleportProbability,
	}
	miner.Walker = *walker.NewWalker(conf, musk.MakeMaxUniformWalk(Next, miner))
	return miner
}

func (w *Walker) Mine(dt lattice.DataType, rptr miners.Reporter, fmtr lattice.Formatter) error {
	errors.Logf("INFO", "Customize creation")

	pConf := w.Config.Copy()
	pConf.Samples = 1000
	pConf.Unique = false
	premine := walker.NewWalker(pConf, ospace.MakeUniformWalk(.02, false))
	premine.Reject = false
	collector := &reporters.Collector{make([]lattice.Node, 0, 10)}
	uniq, err := reporters.NewUnique(pConf, fmtr, collector, "")
	if err != nil {
		return err
	}
	pRptr := &reporters.Skip{
		Skip:     10,
		Reporter: &reporters.Chain{[]miners.Reporter{reporters.NewLog(fmtr, false, "DEBUG", "premining"), uniq}},
	}

	err = premine.Mine(dt, pRptr, fmtr)
	if err != nil {
		return err
	}

	w.Teleports = collector.Nodes
	errors.Logf("INFO", "teleports %v", len(w.Teleports))

	return (w.Walker).Mine(dt, rptr, fmtr)
}

func Next(ctx interface{}, cur lattice.Node) (lattice.Node, error) {
	w := ctx.(*Walker)
	if ismax, err := cur.Maximal(); err != nil {
		return nil, err
	} else if ismax && w.Dt.Acceptable(cur) {
		w.teleportAllowed = true
		errors.Logf("INFO", "ALLOWING TELEPORTS")
	}
	if w.teleportAllowed && rand.Float64() < w.TeleportProbability {
		w.teleportAllowed = false
		next := w.Teleports[rand.Intn(len(w.Teleports))]
		errors.Logf("INFO", "TELEPORT\n    from %v\n      to %v", cur, next)
		return next, nil
	}
	return musk.Next(ctx, cur)
}
