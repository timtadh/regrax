//go:generate fs2-generic --output=wrapper.go --package-name=subgraph_embedding bptree --key-type=*github.com/timtadh/sfp/types/digraph/subgraph/SubGraph --key-serializer=SerializeSubGraph --key-deserializer=DeserializeSubGraph --value-type=*github.com/timtadh/sfp/types/digraph/subgraph/Embedding --value-serializer=SerializeEmbedding --value-deserializer=DeserializeEmbedding

package subgraph_embedding

import (
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

func SerializeSubGraph(sg *subgraph.SubGraph) []byte {
	return sg.Serialize()
}

func DeserializeSubGraph(bytes []byte) *subgraph.SubGraph {
	sg, err := subgraph.LoadSubGraph(bytes)
	if err != nil {
		panic(err)
	}
	return sg
}

func SerializeEmbedding(emb *subgraph.Embedding) []byte {
	return emb.Serialize()
}

func DeserializeEmbedding(bytes []byte) *subgraph.Embedding {
	emb, err := subgraph.LoadEmbedding(bytes)
	if err != nil {
		panic(err)
	}
	return emb
}
