package col

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack[int](nil)
	stack.Push(5)
	stack.Push(6)
	stack.Push(7)
	stack.Push(8)
	fullstack := stack.String()
	assert.Equal(uint64(4), stack.length)
	assert.Equal("HEAD -> 8 -> 7 -> 6 -> 5 -> END", fullstack)
}

func TestPush_strings(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack[string](nil)
	stack.Push("t")
	stack.Push("e")
	stack.Push("s")
	stack.Push("t")
	fullstack := stack.String()
	assert.Equal(uint64(4), stack.length)
	assert.Equal("HEAD -> t -> s -> e -> t -> END", fullstack)
}

func TestPop(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack[int](nil)
	stack.Push(5)
	stack.Push(6)
	stack.Push(7)
	stack.Push(8)
	Popped := stack.Pop()
	assert.Equal(8, *Popped)
	assert.Equal(uint64(3), stack.length)
	Popped = stack.Pop()
	assert.Equal(7, *Popped)
	fullstack := stack.String()
	assert.Equal(uint64(2), stack.length)
	assert.Equal("HEAD -> 6 -> 5 -> END", fullstack)
}

func TestIsZero(t *testing.T) {
	assert := assert.New(t)
	assert.True(IsZero(0))
	zero := *new(int)
	assert.True(IsZero(zero))
}

func TestIsNotZero(t *testing.T) {
	assert := assert.New(t)
	assert.False(IsZero(1))
	assert.True(!IsZero(1))
}
