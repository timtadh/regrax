package digraph

type Mode uint64

const (
	MNI Mode = 1 << iota // Min-Image Support
	GIS                  // Greedy Independent Subgraphs
	FIS                  // Fully Independent Subgraphs
	OverlapPruning       // prune embedding search based on parents embeddings
	EmbeddingPruning     // prune embedding search based on unsupported partial embeddings found during parent search
	ExtensionPruning     // prune extensions based on whether the extensions was unsupported by the parent
	ExtFromEmb           // extend the lattice node from its embeddings
	ExtFromFreqEdges     // extend the lattice node from the frequent edges
	Caching              // enable caching layer (not good for complete mining)
)
