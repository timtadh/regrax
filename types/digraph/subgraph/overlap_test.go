package subgraph

import "testing"

func TestOverlap(t *testing.T) {
	/*
	G, _, sg, indices := graph(t)
	t.Log(sg.Pretty(G.Colors))

	o := sg.FindVertexEmbeddings(indices, 2)
	if o == nil {
		t.Fatal("did not find a supported vertex embedding")
	}
	t.Logf("%v", o.Pretty(G.Colors))
	expected :=  "{6:6}(black:{0, 1})(black:{0, 1})(red:{2, 5})(red:{2, 5})(red:{3, 4})(red:{3, 4})[4->2:][5->3:][0->5:][1->4:][0->2:][1->3:]"
	if o.Pretty(G.Colors) != expected {
		t.Errorf("incorrect overlap")
	}
	*/
}

func TestOverlapEmbeddings(t *testing.T) {
	/*
	G, _, sg, indices := graph(t)
	t.Log(sg.Pretty(G.Colors))

	o := sg.FindVertexEmbeddings(indices, 2)
	if o == nil {
		t.Fatal("did not find a supported vertex embedding")
	}
	t.Logf("%v", o.Pretty(G.Colors))
	expected :=  "{6:6}(black:{0, 1})(black:{0, 1})(red:{2, 5})(red:{2, 5})(red:{3, 4})(red:{3, 4})[4->2:][5->3:][0->5:][1->4:][0->2:][1->3:]"
	if o.Pretty(G.Colors) != expected {
		t.Errorf("incorrect overlap")
	}

	embs := o.SupportedEmbeddings(indices)
	for _, emb := range embs {
		t.Log(emb)
	}
	*/
}
