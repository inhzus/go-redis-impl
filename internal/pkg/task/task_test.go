package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdTask_Task(t *testing.T) {
	tsk := &CmdTask{}
	assert.Equal(t, tsk, tsk.Task())
}

func TestModTask_Task(t *testing.T) {
	tsk := &ModTask{}
	assert.Equal(t, tsk, tsk.Task())
}
