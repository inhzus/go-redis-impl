package label

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToStr(t *testing.T) {
	assert.Equal(t, "", ToStr())
	assert.Equal(t, "string/bulked/array/integer/error/<unknown type>",
		ToStr(String, Bulked, Array, Integer, Error, '0'))
}
