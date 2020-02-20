package load_balancer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWrrNext(t *testing.T) {
	selector := WrrSelector{}
	selector.AddNode(1, 1)
	selector.AddNode(2, 2)
	tryCount := 3
	selected := make([]int, tryCount)
	for i := 0; i < len(selected); i++ {
		item, err := selector.Next()
		assert.True(t, err == nil)
		itemValue, ok := item.(int)
		assert.True(t, ok)
		selected[i] = itemValue
	}
	assert.Equal(t, selected[0], 2)
	assert.Equal(t, selected[1], 1)
	assert.Equal(t, selected[2], 2)
}
