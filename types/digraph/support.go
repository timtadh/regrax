package digraph

import (
	"encoding/json"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/set"
	"github.com/timtadh/data-structures/types"
	"github.com/timtadh/goiso"
)

import (
	"github.com/timtadh/sfp/types/digraph/support"
)

type Supported func(dt *Digraph, sgs []*goiso.SubGraph) ([]*goiso.SubGraph, error)

func MinImgSupported(dt *Digraph, sgs []*goiso.SubGraph) ([]*goiso.SubGraph, error) {
	return support.MinImgSupported(sgs)
}

func MaxIndepSupported(dt *Digraph, sgs []*goiso.SubGraph) ([]*goiso.SubGraph, error) {
	return support.MaxIndepSupported(sgs)
}

func MakeTxSupported(attrName string) Supported {
	return func(dt *Digraph, sgs []*goiso.SubGraph) ([]*goiso.SubGraph, error) {
		sgs, err := support.MinImgSupported(sgs)
		if err != nil {
			return nil, err
		}
		supported := make([]*goiso.SubGraph, 0, len(sgs))
		seen := set.NewSortedSet(len(sgs))
		for _, sg := range sgs {
			if len(sg.V) <= 0 {
				supported = append(supported, sg)
				continue
			}
			vid := int32(sg.V[0].Id)
			err := dt.NodeAttrs.DoFind(vid, func(_ int32, attrs map[string]interface{}) error {
				attr, has := attrs[attrName]
				if !has {
					return errors.Errorf("subgraph %v vertex %v did not have attr %v in attrs %v", sg, sg.V[0], attrName, attrs)
				}
				switch a := attr.(type) {
				case json.Number:
					if n, err := a.Int64(); err != nil {
						return errors.Errorf("type float is not suppported as an attr value for tx id %v %v for %v", attrName, attr, sg)
					} else {
						if !seen.Has(types.Int64(n)) {
							seen.Add(types.Int64(n))
							supported = append(supported, sg)
						}
					}
				case map[string]interface{}:
					return errors.Errorf("type map[string]interface{} is not suppported as an attr value for tx id")
				case []interface{}:
					return errors.Errorf("type []interface{} is not suppported as an attr value for tx id")
				case bool:
					return errors.Errorf("type bool is not suppported as an attr value for tx id")
				case nil:
					return errors.Errorf("type nil is not suppported as an attr value for tx id")
				case string:
					if !seen.Has(types.String(a)) {
						seen.Add(types.String(a))
						supported = append(supported, sg)
					}
				default:
					return errors.Errorf("unexpected type %T for %v in %v", attr, attr, sg)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
		return supported, nil
	}
}
