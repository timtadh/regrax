package reporters

import (
	"fmt"
	"os"
	"encoding/json"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/exc"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/types/digraph"
)

type clusterNode struct {
	n     lattice.Node
	items types.Set
}

func newClusterNode(n lattice.Node, attr string) (*clusterNode, error) {
	items, err := itemset(n, attr)
	if err != nil {
		return nil, err
	}
	cn := &clusterNode{n, items}
	return cn, nil
}

func (a *clusterNode) similarity(b *clusterNode) float64 {
	i, err := a.items.Intersect(b.items)
	exc.ThrowOnError(err)
	inter := float64(i.Size())
	return 1.0 - (inter / (float64(a.items.Size()) + float64(b.items.Size()) - inter))
}

type cluster []*clusterNode

type DbScan struct {
	clusters   []cluster
	config     *config.Config
	fmtr       lattice.Formatter
	filename   string
	attr       string
	epsilon    float64
}

func NewDbScan(c *config.Config, fmtr lattice.Formatter, filename string, attr string, epsilon float64) (*DbScan, error) {
	r := &DbScan{
		config:    c,
		fmtr:      fmtr,
		filename:  filename,
		attr:      attr,
		epsilon:   epsilon,
	}
	return r, nil
}

func (r *DbScan) Report(n lattice.Node) error {
	cn, err := newClusterNode(n, r.attr)
	if err != nil {
		return err
	}
	errors.Logf("DBSCAN", "items %v", cn.items)
	for i := range r.clusters {
		for _, b := range r.clusters[i] {
			if cn.similarity(b) <= r.epsilon {
				r.clusters[i] = append(r.clusters[i], cn)
				return nil
			}
		}
	}
	r.clusters = append(r.clusters, cluster{cn})
	return nil
}

func (r *DbScan) Close() error {
	f, err := os.Create(r.config.OutputFile(r.filename))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")
	for i, cluster := range r.clusters {
		for _, cn := range cluster {
			x := map[string]interface{}{
				"cluster": i,
				"name": r.fmtr.PatternName(cn.n),
				"items": fmt.Sprintf("%v", cn.items),
			}
			err := enc.Encode(x)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func itemset(node lattice.Node, attr string) (types.Set, error) {
	switch n := node.(type) {
	case *digraph.EmbListNode:
		return digraphItemset(n, attr)
	default:
		return nil, errors.Errorf("DBSCAN does not yet support %T", n)
	}
}

func digraphItemset(n *digraph.EmbListNode, attr string) (types.Set, error) {
	dt := n.Dt
	embs, err := n.Embeddings()
	if err != nil {
		return nil, err
	}
	s := set.NewSortedSet(len(embs))
	for _, emb := range embs {
		for _, vid := range emb.Ids {
			err := dt.NodeAttrs.DoFind(
				int32(vid),
				func(_ int32, attrs map[string]interface{}) error {
					if val, has := attrs[attr]; has {
						switch v := val.(type) {
						case string:
							s.Add(types.String(v))
						case int:
							s.Add(types.Int(v))
						default:
							return errors.Errorf("DBSCAN does not yet support attr type %T", val)
						}
					}
					return nil
				})
			if err != nil {
				return nil, err
			}
		}
	}
	return s, nil
}

