package snowflakes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSnowflake(t *testing.T) {
	assert := assert.New(t)

	val, err := NewSnowflake()

	assert.NoError(err, "Must not return an error")

	assert.Len(val, 13, "Snowflake must have correct length")
}

func TestNewRawflake(t *testing.T) {
	val := NewRawFlake()

	if val.Int64() == 0 {
		t.Error("Flake was 0")
		t.FailNow()
	}
}

func BenchmarkNewSnowflake(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = NewSnowflake()
	}
}
