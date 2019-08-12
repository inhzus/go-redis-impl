package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	KeyFmt := "a_%d"
	for i := 0; i < 20; i++ {
		Set(fmt.Sprintf(KeyFmt, i), i, time.Second)
	}
	for i := 0; i < 20; i++ {
		val := Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val.(int), i)
	}
	for i := 0; i < 10; i++ {
		Set(fmt.Sprintf(KeyFmt, i), i+1, time.Second/2)
	}
	for i := 0; i < 10; i++ {
		val := Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val.(int), i+1)
	}
	<-time.After(time.Second / 2)
	for i := 0; i < 10; i++ {
		val := Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val, nil)
	}
	<-time.After(time.Second / 2)
	for i := 11; i < 20; i++ {
		val := Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val, nil)
	}
}

func TestGet(t *testing.T) {
	KeyFmt := "a_%d"
	for i := 0; i < 20; i++ {
		Set(fmt.Sprintf(KeyFmt, i), i, 0)
	}
	assert.Equal(t, Get("a_15"), 15)
	for i := 0; i < 20; i++ {
		Set(fmt.Sprintf(KeyFmt, i), i, time.Millisecond)
	}
	<-time.After(time.Millisecond)
	assert.Equal(t, Get("a_13"), nil)
	for i := 0; i < 20; i++ {
		Set(fmt.Sprintf(KeyFmt, i), i, time.Millisecond)
	}
	<-time.After(time.Millisecond)
	Set("a_15", 15, 0)
	assert.Equal(t, Get("a_15"), 15)
}

