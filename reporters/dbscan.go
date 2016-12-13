package reporters

import (
	"fmt"
	"os"
	"encoding/json"
	"math"
	"math/rand"
	"strings"
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
	labels   types.Set
}

func newClusterNode(fmtr lattice.Formatter, n lattice.Node, attr string) (*clusterNode, error) {
	items, err := itemset(n, attr)
	if err != nil {
		return nil, err
	}
	labels, err := labelset(n)
	if err != nil {
		return nil, err
	}
	cn := &clusterNode{n.Pattern(), fmtr.PatternName(n), items, labels}
	return cn, nil
}

func structureSimilarity(a, b *clusterNode) float64 {
	s := a.pattern.Distance(b.pattern)
	return s
}

func attrSimilarity(a, b *clusterNode) float64 {
	return jaccardSetSimilarity(a.items, b.items)
}

func labelSimilarity(a, b *clusterNode) float64 {
	return jaccardSetSimilarity(a.labels, b.labels)
}

func jaccardSetSimilarity(a, b types.Set) float64 {
	i, err := a.Intersect(b)
	exc.ThrowOnError(err)
	inter := float64(i.Size())
	return 1.0 - (inter / (float64(a.Size()) + float64(b.Size()) - inter))
}

type cluster []*clusterNode

func correlation(clusters []cluster, metric func(a, b *clusterNode) float64) float64 {
	var totalDist float64
	var totalIncidence float64
	var totalItems float64
	for x, X := range clusters {
		for i := 0; i < len(X); i++ {
			for y := x; y < len(clusters); y++ {
				Y := clusters[y]
				for j := i + 1; j < len(Y); j++ {
					totalDist += metric(X[i], Y[j])
					if x == y {
						totalIncidence++
					}
					totalItems++
				}
			}
		}
	}
	meanDist := totalDist/totalItems
	meanIncidence := totalIncidence/totalItems
	var sumOfSqDist float64
	var sumOfSqIncidence float64
	var sumOfProduct float64
	for x, X := range clusters {
		for i := 0; i < len(X); i++ {
			for y := x; y < len(clusters); y++ {
				Y := clusters[y]
				for j := i + 1; j < len(Y); j++ {
					dist := metric(X[i], Y[j])
					var incidence float64
					if x == y {
						incidence = 1.0
					}
					distDiff := (dist - meanDist)
					incidenceDiff := (incidence - meanIncidence)
					sumOfSqDist += distDiff*distDiff
					sumOfSqIncidence += incidenceDiff*incidenceDiff
					sumOfProduct += distDiff*incidenceDiff
				}
			}
		}
	}
	return sumOfProduct / (sumOfSqDist * sumOfSqIncidence)
}

func intradist(clusters []cluster, metric func(a, b *clusterNode) float64) float64 {
	var totalDist float64
	for _, X := range clusters {
		var dist float64
		for i := 0; i < len(X); i++ {
			for j := i + 1; j < len(X); j++ {
				dist += metric(X[i], X[j])
			}
		}
		if len(X) > 0 {
			totalDist += dist/float64(len(X))
		}
	}
	return totalDist
}

func interdist(clusters []cluster, metric func(a, b *clusterNode) float64) float64 {
	var totalDist float64
	for x, X := range clusters {
		var dist float64
		for i := 0; i < len(X); i++ {
			for y := x; y < len(clusters); y++ {
				Y := clusters[y]
				var to float64
				for j := i + 1; j < len(Y); j++ {
					if y == x {
						continue
					}
					for j := 0; j < len(Y); j++ {
						to += metric(X[i], Y[j])
					}
				}
				if len(Y) > 0 {
					dist += to/float64(len(Y))
				}
			}
		}
		if len(X) > 0 {
			totalDist += dist/float64(len(X))
		}
	}
	return totalDist
}

func noNan(x float64) interface{} {
	if math.IsNaN(x) {
		return "nan"
	}
	return x
}

type DbScan struct {
	clusters   []cluster
	items      int
	config     *config.Config
	fmtr       lattice.Formatter
	clustersName, metricsName   string
	attr       string
	epsilon    float64
	gamma      float64
}

func NewDbScan(c *config.Config, fmtr lattice.Formatter, clusters, metrics string, attr string, epsilon, gamma float64) (*DbScan, error) {
	r := &DbScan{
		config:    c,
		fmtr:      fmtr,
		clustersName:  clusters,
		metricsName:  metrics,
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
	r.items++
	r.clusters = add(r.clusters, cn, r.epsilon, labelSimilarity)
	return nil
}

func (r *DbScan) Close() error {
	f, err := os.Create(r.config.OutputFile(r.clustersName))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")
	random := make([]cluster, rand.Intn(int(float64(len(r.clusters))*1.2)) + 2)
	for i := range r.clusters {
		groups := make([]cluster, 0, 10)
		for _, cn := range r.clusters[i] {
			// groups = add(groups, cn, r.gamma, structureSimilarity)
			groups = add(groups, cn, r.epsilon, attrSimilarity)
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
				n := rand.Intn(len(random))
				random[n] = append(random[n], cn)
			}
		}
	}
	return r.metrics(random)
}

func (r *DbScan) metrics(random []cluster) error {
	f, err := os.Create(r.config.OutputFile(r.metricsName))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")
	if len(r.clusters) >= r.items || len(r.clusters) <= 1 {
		x := map[string]interface{} {
			"items": r.items,
			"clusters": len(r.clusters),
		}
		return enc.Encode(x)
	}
	intraLabel := intradist(r.clusters, labelSimilarity)
	interLabel := interdist(r.clusters, labelSimilarity)
	intraAttr := intradist(r.clusters, attrSimilarity)
	interAttr := interdist(r.clusters, attrSimilarity)
	intraLabelRand := intradist(random, labelSimilarity)
	interLabelRand := interdist(random, labelSimilarity)
	intraAttrRand := intradist(random, attrSimilarity)
	interAttrRand := interdist(random, attrSimilarity)
	stderr := func(a, b float64) float64 {
		return math.Sqrt((a-b)*(a-b))
	}
	x := map[string]interface{} {
		"items": r.items,
		"cluster-metrics": map[string]interface{} {
			"count": len(r.clusters),
			"attr-correlation": noNan(correlation(r.clusters, attrSimilarity)),
			"label-correlation": noNan(correlation(r.clusters, labelSimilarity)),
			"label-intra-distance": noNan(intraLabel),
			"label-inter-distance": noNan(interLabel),
			"label-distance-ratio": noNan(intraLabel/interLabel),
			"attr-intra-distance": noNan(intraAttr),
			"attr-inter-distance": noNan(interAttr),
			"attr-distance-ratio": noNan(intraAttr/interAttr),
		},
		"random-metrics": map[string]interface{} {
			"count": len(random),
			"attr-correlation": noNan(correlation(random, attrSimilarity)),
			"label-correlation": noNan(correlation(random, labelSimilarity)),
			"label-intra-distance": noNan(intraLabelRand),
			"label-inter-distance": noNan(interLabelRand),
			"label-distance-ratio": noNan(intraLabelRand/interLabelRand),
			"attr-intra-distance": noNan(intraAttrRand),
			"attr-inter-distance": noNan(interAttrRand),
			"attr-distance-ratio": noNan(intraAttrRand/interAttrRand),
		},
		"standard-error":map[string]interface{} {
			"count": noNan(stderr(float64(len(r.clusters)), float64(len(random)))),
			"attr-correlation": noNan(stderr(correlation(r.clusters, attrSimilarity), correlation(random, attrSimilarity))),
			"label-correlation": noNan(stderr(correlation(r.clusters, labelSimilarity), correlation(random, labelSimilarity))),
			"label-intra-distance": noNan(stderr(intraLabel, intraLabelRand)),
			"label-inter-distance": noNan(stderr(interLabel, interLabelRand)),
			"label-distance-ratio": noNan(stderr(intraLabel/interLabel, intraLabelRand/interLabelRand)),
			"attr-intra-distance": noNan(stderr(intraAttr, intraAttrRand)),
			"attr-inter-distance": noNan(stderr(interAttr, interAttrRand)),
			"attr-distance-ratio": noNan(stderr(intraAttr/interAttr, intraAttrRand/interAttrRand)),
		},
	}
	err = enc.Encode(x)
	if err != nil && strings.Contains(err.Error(), "NaN") {
		x := map[string]interface{} {
			"items": r.items,
			"clusters": len(r.clusters),
		}
		return enc.Encode(x)
	} else if err != nil {
		return err
	}
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
	}
	if false {
		errors.Logf("DBSCAN", "%v %v %v", min_sim, cn.pattern, min_item.pattern)
	}
	clusters[min_near] = append(clusters[min_near], cn)
	prev := -1
	for x, next := near.ItemsInReverse()(); next != nil; x, next = next() {
		cur := int(x.(types.Int))
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

func labelset(node lattice.Node) (types.Set, error) {
	switch n := node.(type) {
	case *digraph.EmbListNode:
		return digraphLabelset(n)
	default:
		return nil, errors.Errorf("DBSCAN does not yet support %T", n)
	}
}

func digraphLabelset(n *digraph.EmbListNode) (types.Set, error) {
	p := n.Pat
	s := set.NewSortedSet(len(p.V) + len(p.E))
	for i := range p.V {
		s.Add(types.Int(p.V[i].Color))
	}
	for i := range p.E {
		s.Add(types.Int(p.E[i].Color))
	}
	return s, nil
}


