package core

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type testType struct {
	Val int
}

func (t *testType) ID() string {
	return uuid.NewString()
}

func TestAddNode(t *testing.T) {
	graph := NewGraph()
	id1, err := graph.AddVertex(&testType{Val: 1})
	assert.NoError(t, err)
	id2, err := graph.AddVertex(&testType{Val: 2})
	assert.NoError(t, err)

	assert.NotNil(t, graph.nodes[id1])
	assert.NotNil(t, graph.nodes[id2])

	assert.Equal(t, 1, graph.nodes[id1].value.(*testType).Val)
	assert.Equal(t, 2, graph.nodes[id2].value.(*testType).Val)
}

func TestAddEdge(t *testing.T) {
	graph := NewGraph()
	id1, err := graph.AddVertex(&testType{Val: 1})
	assert.NoError(t, err)
	id2, err := graph.AddVertex(&testType{Val: 2})
	assert.NoError(t, err)
	id3, err := graph.AddVertex(&testType{Val: 3})
	assert.NoError(t, err)

	err = graph.AddEdge(id1, id2)
	assert.NoError(t, err)
	err = graph.AddEdge(id1, id3)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(graph.nodes[id1].children))
	assert.Equal(t, 0, len(graph.nodes[id2].children))
	assert.Equal(t, 0, len(graph.nodes[id3].children))
}

func TestAddEdge_CycleDetect(t *testing.T) {
	graph := NewGraph()
	id1, err := graph.AddVertex(&testType{Val: 1})
	assert.NoError(t, err)
	id2, err := graph.AddVertex(&testType{Val: 2})
	assert.NoError(t, err)
	id3, err := graph.AddVertex(&testType{Val: 3})
	assert.NoError(t, err)

	err = graph.AddEdge(id1, id2)
	assert.NoError(t, err)
	err = graph.AddEdge(id1, id3)
	assert.NoError(t, err)
	err = graph.AddEdge(id2, id2)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "Cycle detected when trying to add edge.")
}

type testTraverse struct {
	path string
}

func (t *testTraverse) Do(val *Traverser) {
	if t.path == "" {
		t.path = fmt.Sprintf("%d", ((*val).node.value.(*testType)).Val)
	} else {
		t.path = fmt.Sprintf("%s,%d", t.path, ((*val).node.value.(*testType)).Val)
	}
}

func TestTraverseBF(t *testing.T) {
	graph := NewGraph()
	id1, _ := graph.AddVertex(&testType{Val: 1})
	id2, _ := graph.AddVertex(&testType{Val: 2})
	id3, _ := graph.AddVertex(&testType{Val: 3})

	err := graph.AddEdge(id1, id2)
	assert.NoError(t, err)
	err = graph.AddEdge(id1, id3)
	assert.NoError(t, err)

	actualTraverse := testTraverse{path: ""}
	graph.TraverseBF(&actualTraverse)
	assert.Equal(t, "1,2,3", actualTraverse.path)
}

func TestTraverseBFFrom(t *testing.T) {
	graph := NewGraph()
	id1, _ := graph.AddVertex(&testType{Val: 1})
	id2, _ := graph.AddVertex(&testType{Val: 2})
	id3, _ := graph.AddVertex(&testType{Val: 3})
	id4, _ := graph.AddVertex(&testType{Val: 4})

	err := graph.AddEdge(id1, id2)
	assert.NoError(t, err)
	err = graph.AddEdge(id1, id3)
	assert.NoError(t, err)
	err = graph.AddEdge(id3, id4)
	assert.NoError(t, err)

	actualTraverse := testTraverse{path: ""}
	graph.TraverseBFFrom(id3, &actualTraverse)
	assert.Equal(t, "3,4", actualTraverse.path)
}
