package core

import (
	"sort"
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
	path []int
}

// Do appends the visited node's value to the path slice.
func (t *testTraverse) Do(val *Traverser) {
	// The value is asserted to be of *testType to extract the integer.
	nodeValue := ((*val).node.value.(*testType)).Val
	t.path = append(t.path, nodeValue)
}

func TestTraverseBF(t *testing.T) {
	graph := NewGraph()
	id1, _ := graph.AddVertex(&testType{Val: 1})
	id2, _ := graph.AddVertex(&testType{Val: 2})
	id3, _ := graph.AddVertex(&testType{Val: 3})
	id4, _ := graph.AddVertex(&testType{Val: 4})

	// second tree
	id5, _ := graph.AddVertex(&testType{Val: 5})
	id6, _ := graph.AddVertex(&testType{Val: 6})

	err := graph.AddEdge(id1, id2)
	assert.NoError(t, err)
	err = graph.AddEdge(id2, id3)
	assert.NoError(t, err)
	err = graph.AddEdge(id2, id4)
	assert.NoError(t, err)

	err = graph.AddEdge(id5, id6)
	assert.NoError(t, err)

	actualTraverse := testTraverse{}
	graph.TraverseBF(&actualTraverse)
	path := actualTraverse.path

	assert.Len(t, path, 6, "Should have visited all 6 nodes")

	level0 := path[0:2]
	sort.Ints(level0)
	assert.Equal(t, []int{1, 5}, level0, "Level 0 nodes (roots) should be {1, 5}")

	level1 := path[2:4]
	sort.Ints(level1)
	assert.Equal(t, []int{2, 6}, level1, "Level 1 nodes should be {2, 6}")

	level2 := path[4:6]
	sort.Ints(level2)
	assert.Equal(t, []int{3, 4}, level2, "Level 2 nodes should be {3, 4}")
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

	actualTraverse := testTraverse{}
	graph.TraverseBFFrom(id3, &actualTraverse)
	path := actualTraverse.path

	level0 := path[0:2]
	sort.Ints(level0)
	assert.Equal(t, []int{3, 4}, level0, "Level 0 nodes (roots) should be {3, 4}")
}
