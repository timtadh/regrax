package subgraph

import "testing"

func TestOverlap(t *testing.T) {
	G, _, sg, indices := graph(t)
	t.Log(sg.Pretty(G.Colors))

	ve, err := sg.FindVertexEmbeddings(G, indices, 2)
	if err != nil {
		t.Fatal(err)
	}
	if ve == nil {
		t.Fatal("did not find a supported vertex embedding")
	}
}
