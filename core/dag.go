package core

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

type IDInterface interface {
	ID() string
}

type traverseAction interface {
	Do(node *node)
}

// node represents a node in the graph
type node struct {
	value    interface{}
	id       string
	children []*node
}

func (n *node) IsLeaf() bool {
	return len(n.children) == 0
}

// AcyclicGraph represents a directed acyclic AcyclicGraph
type AcyclicGraph struct {
	nodes map[string]*node
	muDAG sync.RWMutex
}

// NewGraph creates a new graph
func NewGraph() *AcyclicGraph {
	return &AcyclicGraph{nodes: make(map[string]*node)}
}

// AddVertex adds a node to the graph
func (g *AcyclicGraph) AddVertex(value interface{}) (string, error) {
	g.muDAG.Lock()
	defer g.muDAG.Unlock()
	id := value.(IDInterface).ID()
	if _, exists := g.nodes[id]; !exists {
		g.nodes[id] = &node{value: value, id: id}
		return id, nil
	}
	return "", errors.New("Node with same value already exists in graph")
}

// AddEdge adds a directed edge from one node to another
func (g *AcyclicGraph) AddEdge(from, to string) error {
	g.muDAG.Lock()
	defer g.muDAG.Unlock()
	fromNode := g.nodes[from]
	toNode := g.nodes[to]
	if fromNode != nil && toNode != nil {
		fromNode.children = append(fromNode.children, toNode)
	}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	if g.hasCycleHelper(fromNode, visited, recStack) {
		fromNode.children = fromNode.children[:len(fromNode.children)-1]
		return errors.New("cycle detected")
	}
	return nil
}

func (g *AcyclicGraph) GetVertex(id string) (*node, error) {
	g.muDAG.Lock()
	defer g.muDAG.Unlock()
	if node, ok := g.nodes[id]; !ok {
		return nil, errors.New("no vertex with specified id")
	} else {
		return node, nil
	}
}

// Traverse the graph breadt-first from a specified start node ID
func (g *AcyclicGraph) TraverseBFFrom(start string, tAction traverseAction) {
	g.muDAG.Lock()
	defer g.muDAG.Unlock()
	startNode, exists := g.nodes[start]
	if !exists {
		fmt.Println("Start node not found in the graph")
		return
	}

	visited := make(map[string]bool)
	queue := list.New()
	queue.PushBack(startNode)
	visited[startNode.id] = true

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		node := element.Value.(*node)
		tAction.Do(node)

		for _, child := range node.children {
			if !visited[child.id] {
				queue.PushBack(child)
				visited[child.id] = true
			}
		}
	}
}

// Traverse the entire graph breadth-first
func (g *AcyclicGraph) TraverseBF(tAction traverseAction) {
	g.muDAG.Lock()
	defer g.muDAG.Unlock()
	visited := make(map[string]bool)
	queue := list.New()

	for _, id := range g.findRoots() {
		startNode := g.nodes[id]
		if !visited[startNode.id] {
			queue.PushBack(startNode)
			visited[startNode.id] = true

			for queue.Len() > 0 {
				element := queue.Front()
				queue.Remove(element)
				node := element.Value.(*node)
				tAction.Do(node)

				for _, child := range node.children {
					if !visited[child.id] {
						queue.PushBack(child)
						visited[child.id] = true
					}
				}
			}
		}
	}
}

// hasCycleHelper is a utility function to check for cycles in the graph
func (g *AcyclicGraph) hasCycleHelper(node *node, visited map[string]bool, recStack map[string]bool) bool {
	if recStack[node.id] {
		return true
	}
	if visited[node.id] {
		return false
	}

	visited[node.id] = true
	recStack[node.id] = true

	for _, child := range node.children {
		if g.hasCycleHelper(child, visited, recStack) {
			return true
		}
	}

	recStack[node.id] = false
	return false
}

// hasCycle checks if the graph has a cycle
// func (g *acyclicGraph) hasCycle() bool {
// 	visited := make(map[string]bool)
// 	recStack := make(map[string]bool)

// 	for _, node := range g.nodes {
// 		if !visited[node.id] {
// 			if g.hasCycleHelper(node, visited, recStack) {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

func (g *AcyclicGraph) calculateInDegrees() map[string]int {
	inDegree := make(map[string]int)

	// Initialize in-degree of all nodes to 0
	for key, _ := range g.nodes {
		inDegree[key] = 0
	}

	// Compute in-degree of each node
	for _, node := range g.nodes {
		for _, child := range node.children {
			inDegree[child.id]++
		}
	}

	return inDegree
}

func (g *AcyclicGraph) findRoots() []string {
	inDegree := g.calculateInDegrees()
	var roots []string

	for id, degree := range inDegree {
		if degree == 0 {
			roots = append(roots, id)
		}
	}

	return roots
}
