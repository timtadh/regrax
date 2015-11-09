package itemset

import "testing"
import "github.com/stretchr/testify/assert"

func TestHello(x *testing.T) {
	t := assert.New(x)
	t.Equal("hello", string([]byte("hello")), "wizard %v", 1)
}

