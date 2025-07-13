package core

import (
	"container/list"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type IDInterface interface {
	ID() string
}

type Traverser struct {
	node *node
}

func (t *Traverser) GetVertex() *node {
	return t.node
}

type traverseAction interface {
	Do(node *Traverser)
}

// node represents a node in the graph
type node struct {
	value    interface{}
	id       string
	children []*node
	isRoot   bool
}

func (n *node) IsLeaf() bool {
	return len(n.children) == 0
}

func (n *node) IsRoot() bool {
	return n.isRoot
}

func (n *node) Value() interface{} {
	return n.value
}

// AcyclicGraph represents a directed acyclic AcyclicGraph
type AcyclicGraph struct {
	nodes map[string]*node
	size  int
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
	fromNode, ok := g.nodes[from]
	if !ok {
		return errors.New("from vertex does not exist")
	}
	toNode, ok := g.nodes[to]
	if !ok {
		return errors.New("to vertex does not exist")
	}
	if fromNode.hasEdgeTo(to) {
		return &EdgeExistsError{}
	}
	if fromNode != nil && toNode != nil {
		fromNode.children = append(fromNode.children, toNode)
	}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	if g.hasCycleHelper(fromNode, visited, recStack) {
		fromNode.children = fromNode.children[:len(fromNode.children)-1]
		return &EdgeCycleError{}
	}
	g.size++
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

// Traverse the graph breadth-first from a specified start node ID
func (g *AcyclicGraph) TraverseBFFrom(start string, tAction traverseAction) {
	g.muDAG.Lock()
	defer g.muDAG.Unlock()
	startNode, exists := g.nodes[start]
	if !exists {
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
		tAction.Do(&Traverser{node: node})

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
		startNode.isRoot = true
		if !visited[startNode.id] {
			queue.PushBack(startNode)
			visited[startNode.id] = true
		}
	}

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		node := element.Value.(*node)
		tAction.Do(&Traverser{node: node})

		for _, child := range node.children {
			if !visited[child.id] {
				queue.PushBack(child)
				visited[child.id] = true
			}
		}
	}
}

// TraverseTopological visits nodes in a topological order (parents before children).
// It takes a traverseAction to perform on each node as it's visited.
// It returns an error if the graph contains a cycle, as a topological
// sort is not possible.
func (g *AcyclicGraph) TraverseTopological(tAction traverseAction) error {
	g.muDAG.RLock()
	defer g.muDAG.RUnlock()

	inDegree := g.calculateInDegrees()

	// Initialize a queue with all nodes that have an in-degree of 0.
	queue := list.New()
	for id, degree := range inDegree {
		if degree == 0 {
			node := g.nodes[id]
			node.isRoot = true // Set the root flag, similar to TraverseBF
			queue.PushBack(node)
		}
	}

	visitedCount := 0
	for queue.Len() > 0 {
		// Dequeue the next node in the topological order.
		element := queue.Front()
		queue.Remove(element)
		node := element.Value.(*node)

		// Perform the user-defined action on the node.
		tAction.Do(&Traverser{node: node})
		visitedCount++

		// For each child of the visited node, decrement its in-degree.
		for _, child := range node.children {
			inDegree[child.id]--
			// If a child's in-degree drops to 0, all its parents have
			// been visited, so it's ready to be added to the queue.
			if inDegree[child.id] == 0 {
				queue.PushBack(child)
			}
		}
	}

	// If not all nodes were visited, the graph must have a cycle.
	if visitedCount != len(g.nodes) {
		return errors.New("graph has a cycle, topological traversal not possible")
	}

	return nil
}

// TraverseTopologicalFrom performs a topological sort on the subgraph of all nodes
// reachable from a given start node ID. It executes the provided traverseAction
// on each node in the subgraph in topological order.
func (g *AcyclicGraph) TraverseTopologicalFrom(start string, tAction traverseAction) error {
	g.muDAG.RLock()
	defer g.muDAG.RUnlock()

	startNode, exists := g.nodes[start]
	if !exists {
		return errors.New("start node does not exist")
	}

	// Step 1: Find all nodes reachable from the start node to define our subgraph.
	subgraphNodes := make(map[string]*node)
	q := list.New()
	q.PushBack(startNode)
	visitedForSubgraph := make(map[string]bool)
	visitedForSubgraph[start] = true

	for q.Len() > 0 {
		curr := q.Front().Value.(*node)
		q.Remove(q.Front())
		subgraphNodes[curr.id] = curr

		for _, child := range curr.children {
			if !visitedForSubgraph[child.id] {
				visitedForSubgraph[child.id] = true
				q.PushBack(child)
			}
		}
	}

	subgraphInDegree := make(map[string]int)
	for id := range subgraphNodes {
		subgraphInDegree[id] = 0 // Initialize all subgraph nodes
	}
	for _, node := range subgraphNodes {
		for _, child := range node.children {
			// Important: Only count edges where the child is also in our subgraph.
			if _, ok := subgraphNodes[child.id]; ok {
				subgraphInDegree[child.id]++
			}
		}
	}

	topoQueue := list.New()
	for id, degree := range subgraphInDegree {
		if degree == 0 {
			topoQueue.PushBack(subgraphNodes[id])
		}
	}

	visitedCount := 0
	for topoQueue.Len() > 0 {
		node := topoQueue.Front().Value.(*node)
		topoQueue.Remove(topoQueue.Front())

		tAction.Do(&Traverser{node: node})
		visitedCount++

		for _, child := range node.children {
			if _, ok := subgraphNodes[child.id]; ok {
				subgraphInDegree[child.id]--
				if subgraphInDegree[child.id] == 0 {
					topoQueue.PushBack(child)
				}
			}
		}
	}

	if visitedCount != len(subgraphNodes) {
		return errors.New("subgraph has a cycle, topological traversal not possible")
	}

	return nil
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

func (g *AcyclicGraph) GetOrder() int {
	return len(g.nodes)
}

func (g *AcyclicGraph) GetSize() int {
	return g.size
}

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

func (n *node) hasEdgeTo(id string) bool {
	for _, c := range n.children {
		if c.id == id {
			return true
		}
	}
	return false
}

func (g *AcyclicGraph) Print() {
	g.muDAG.RLock()
	defer g.muDAG.RUnlock()

	fmt.Println("--- Graph Structure ---")
	for id, node := range g.nodes {
		if len(node.children) > 0 {
			var childIDs []string
			for _, child := range node.children {
				childIDs = append(childIDs, child.id)
			}
			fmt.Printf("Node %s -> [%s]\n", id, strings.Join(childIDs, ", "))
		} else {
			fmt.Printf("Node %s -> [] (Leaf Node)\n", id)
		}
	}
	fmt.Println("-----------------------")
}
