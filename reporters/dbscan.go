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
	pattern  lattice.Pattern
	name     string
	items    types.Set
}

func newClusterNode(fmtr lattice.Formatter, n lattice.Node, attr string) (*clusterNode, error) {
	items, err := itemset(n, attr)
	if err != nil {
		return nil, err
	}
	cn := &clusterNode{n.Pattern(), fmtr.PatternName(n), items}
	return cn, nil
}

func structureSimilarity(a, b *clusterNode) float64 {
	s := a.pattern.Distance(b.pattern)
	return s
}

func attrSimilarity(a, b *clusterNode) float64 {
	return jaccardSetSimilarity(a.items, b.items)
}

func jaccardSetSimilarity(a, b types.Set) float64 {
	i, err := a.Intersect(b)
	exc.ThrowOnError(err)
	inter := float64(i.Size())
	return 1.0 - (inter / (float64(a.Size()) + float64(b.Size()) - inter))
}

type cluster []*clusterNode

type DbScan struct {
	clusters   []cluster
	config     *config.Config
	fmtr       lattice.Formatter
	filename   string
	attr       string
	epsilon    float64
	gamma      float64
}

func NewDbScan(c *config.Config, fmtr lattice.Formatter, filename string, attr string, epsilon, gamma float64) (*DbScan, error) {
	r := &DbScan{
		config:    c,
		fmtr:      fmtr,
		filename:  filename,
		attr:      attr,
		epsilon:   epsilon,
		gamma:     gamma,
	}
	return r, nil
}

func (r *DbScan) Report(n lattice.Node) error {
	cn, err := newClusterNode(r.fmtr, n, r.attr)
	if err != nil {
		return err
	}
	errors.Logf("DBSCAN", "items %v", cn.items)
	r.clusters = add(r.clusters, cn, r.epsilon, attrSimilarity)
	return nil
}

func add(clusters []cluster, cn *clusterNode, epsilon float64, sim func(a, b *clusterNode) float64) []cluster {
	near := set.NewSortedSet(len(clusters))
	min_near := -1
	min_sim := -1.0
	var min_item *clusterNode = nil
	for i := len(clusters) - 1; i >= 0; i-- {
		for _, b := range clusters[i] {
			s := sim(cn, b)
			if s <= epsilon  {
				near.Add(types.Int(i))
				if min_near == -1 || s < min_sim {
					min_near = i
					min_sim = s
					min_item = b
				}
			}
		}
	}
	if near.Size() <= 0 {
		return append(clusters, cluster{cn})
	} else {
		errors.Logf("DBSCAN", "%v %v %v", min_sim, cn.pattern, min_item.pattern)
		clusters[min_near] = append(clusters[min_near], cn)
		return clusters
	}

	prev := -1
	for x, next := near.ItemsInReverse()(); next != nil; x, next = next() {
		cur := int(x.(types.Int))
		if prev == -1 {
			clusters[cur] = append(clusters[cur], cn)
		}
		if prev >= 0 {
			clusters[cur] = append(clusters[cur], clusters[prev]...)
			clusters = remove(clusters, prev)
		}
		prev = cur
	}
	return clusters
}

func remove(list []cluster, i int) []cluster {
	if i >= len(list) {
		panic(fmt.Errorf("out of range (i (%v) >= len(list) (%v))", i, len(list)))
	} else if i < 0 {
		panic(fmt.Errorf("out of range (i (%v) < 0)", i))
	}
	for ; i < len(list) - 1; i++ {
		list[i] = list[i+1]
	}
	list[len(list)-1] = nil
	return list[:len(list)-1]
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
	for i := range r.clusters {
		groups := make([]cluster, 0, 10)
		for _, cn := range r.clusters[i] {
			groups = add(groups, cn, r.gamma, structureSimilarity)
		}
		for j, cluster := range groups {
			for _, cn := range cluster {
				x := map[string]interface{}{
					"cluster": i,
					"group": j,
					"name": cn.name,
					"items": fmt.Sprintf("%v", cn.items),
				}
				err := enc.Encode(x)
				if err != nil {
					return err
				}
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

