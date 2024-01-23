/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"fmt"
	"reflect"

	"github.com/ShounakA/roids/col"
	"github.com/heimdalr/dag"
)

type (
	serviceGraph struct {
		dag *dag.DAG
	}

	// Dependency visitor. It keeps track of the nodes visited into a stack,
	// so that we can instantiate leaf deps by popping them out.
	depVisiter struct {
		// History of the dependent services visited.
		Hist col.IStack[reflect.Type]
	}

	// Struct to perform a lookup from the search type.
	reverseLookupVisiter struct {
		vertexId   string
		Service    *Service
		searchType reflect.Type
	}
)

// Create a new service graph, with custom pointer functions.
func newServiceGraph(d *dag.DAG) *serviceGraph {
	return &serviceGraph{
		dag: d,
	}
}

// Checks if the provided the vertex (via ID) is a leaf
func (graph *serviceGraph) IsLeafIgnoreError(id string) bool {
	isLeaf, _ := graph.dag.IsLeaf(id)
	return isLeaf
}

// Gets the order of instantiation, by traversing the graph breadth-first
func (graph *serviceGraph) GetInstantiationOrder() col.IStack[reflect.Type] {
	v := depVisiter{Hist: col.NewStack[reflect.Type](nil)}
	graph.dag.BFSWalk(&v)
	return v.Hist
}

// Gets the order of instantiation of the , by traversing the graph breadth-first
func (graph *serviceGraph) GetServiceOrderById(id string) col.IStack[string] {
	chVertex, _, _ := graph.dag.DescendantsWalker(id)
	hist := col.NewStack[string](&id)
	select {
	case vertexId := <-chVertex:
		if vertexId != "" {
			hist.Push(vertexId)
		}
	}
	return hist
}

// Gets the Service struct from the graph by the interface type provided.
func (graph *serviceGraph) GetServiceByType(specType reflect.Type) *Service {
	lookup := &reverseLookupVisiter{searchType: specType, Service: nil}
	graph.dag.BFSWalk(lookup)
	return lookup.Service
}

func (graph *serviceGraph) GetVertex(id string) (*Service, error) {
	vertex, err := graph.dag.GetVertex(id)
	if err != nil {
		return nil, err
	}
	service := vertex.(*Service)
	return service, nil
}

func (graph *serviceGraph) AddVertex(service *Service) error {
	id, err := graph.dag.AddVertex(service)
	service.Id = id
	if err != nil {
		return err
	}
	return nil
}

func (graph *serviceGraph) AddEdge(srcService *Service, depService *Service) error {
	err := graph.dag.AddEdge(srcService.Id, depService.Id)
	if err != nil {
		switch e := err.(type) {
		case dag.EdgeLoopError:
			return NewCircularDependencyError(e, srcService.SpecType)
		case dag.EdgeDuplicateError:
			return NewDuplicateEdgeError(e, srcService.Id, srcService.SpecType)
		default:
			return NewUnknownError(e)
		}
	}
	return nil
}

// Function to clear the services graph
func (graph *serviceGraph) ClearGraph() {
	graph.dag = dag.NewDAG()
}

// Function to show the state of the graph
func (graph *serviceGraph) ShowGraph() {
	fmt.Println(graph.dag.String())
}

// Visit implementation to traverse the entire dependency graph breadth-first
func (pv *depVisiter) Visit(v dag.Vertexer) {
	roids := GetRoids()
	id, value := v.Vertex()
	service := value.(*Service)
	pv.Hist.Push(service.SpecType)
	isLeaf := roids.servicesGraph.IsLeafIgnoreError(id)
	service.isLeaf = isLeaf
}

// Function to lookup vertexId based on spec
func (pv *reverseLookupVisiter) Visit(v dag.Vertexer) {
	id, value := v.Vertex()
	service := value.(*Service)
	if service.SpecType == pv.searchType {
		pv.vertexId = id
		pv.Service = value.(*Service)
		return
	}
}
