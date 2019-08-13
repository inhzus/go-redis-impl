package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItfToBulked(t *testing.T) {
	var val interface{}
	var err error
	val, err = ItfToBulked(nil)
	assert.Nil(t, val)
	assert.Nil(t, err)

	val, err = ItfToBulked([]byte("test"))
	assert.Equal(t, []byte("test"), val)
	assert.Nil(t, err)

	val, err = ItfToBulked(10)
	assert.Equal(t, []byte("10"), val)
	assert.Nil(t, err)

	val, err = ItfToBulked(int64(20))
	assert.Equal(t, []byte("20"), val)
	assert.Nil(t, err)

	val, err = ItfToBulked("txt")
	assert.Equal(t, []byte("txt"), val)
	assert.Nil(t, err)
	val, err = ItfToBulked(3.0)
	assert.Nil(t, val)
	assert.EqualError(t, err, "value cannot cast to bulked")
}

func TestItfToInt(t *testing.T) {
	var val interface{}
	var err error
	val, err = ItfToInt(nil)
	assert.Nil(t, val)
	assert.Nil(t, err)

	val, err = ItfToInt([]byte("13"))
	assert.Equal(t, int64(13), val)
	assert.Nil(t, err)

	val, err = ItfToInt([]byte("13.0"))
	assert.Nil(t, val)
	assert.EqualError(t, err, "value (13.0) cannot cast to int")

	val, err = ItfToInt("100")
	assert.Equal(t, int64(100), val)
	assert.Nil(t, err)

	val, err = ItfToInt("100.0")
	assert.Nil(t, val)
	assert.EqualError(t, err, "value (100.0) cannot cast to int")

	val, err = ItfToInt(int64(200))
	assert.Equal(t, int64(200), val)
	assert.Nil(t, err)

	val, err = ItfToInt(200)
	assert.Nil(t, val)
	assert.EqualError(t, err, "value (200) cannot cast to int")
}
