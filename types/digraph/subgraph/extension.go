package subgraph

type Extension struct {
	Source Vertex
	Target Vertex
	Color int
}

func NewExt(src, targ Vertex, color int) *Extension {
	return &Extension{
		Source: src,
		Target: targ,
		Color: color,
	}
}

