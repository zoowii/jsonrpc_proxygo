package rate_limit

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTokenBucketLimiterTake(t *testing.T) {
	rate := 3
	unit := time.Second
	ul := NewTokenBucketLimiter(rate, unit)
	for i := 0; i < 4; i++ {
		if i < rate {
			assert.True(t, ul.Take())
		} else {
			assert.False(t, ul.Take())
		}
	}
}
