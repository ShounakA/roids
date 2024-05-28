/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ShounakA/roids/col"
	"github.com/ShounakA/roids/core"
	"github.com/heimdalr/dag"
)

type (
	serviceGraph struct {
		dag  *dag.DAG
		dag2 *core.AcyclicGraph
	}

	// Dependency visitor. It keeps track of the nodes visited into a stack,
	// so that we can instantiate leaf deps by popping them out.
	depVisiter struct {
		// History of the dependent services visited.
		Hist col.IStack[string]
	}

	// Struct to perform a lookup from the search type.
	reverseLookupVisiter struct {
		vertexId   string
		Service    *Service
		searchType reflect.Type
	}
)

// Create a new service graph, with custom pointer functions.
func newServiceGraph(d *dag.DAG, d2 *core.AcyclicGraph) *serviceGraph {
	return &serviceGraph{
		dag:  d,
		dag2: d2,
	}
}

// Checks if the provided the vertex (via ID) is a leaf
func (graph *serviceGraph) IsLeafIgnoreError(id string) bool {
	isLeaf, _ := graph.dag.IsLeaf(id)
	return isLeaf
}

// Gets the order of instantiation, by traversing the graph breadth-first
func (graph *serviceGraph) GetInstantiationOrder() col.IStack[string] {
	v := depVisiter{Hist: col.NewStack[string](nil)}
	graph.dag.BFSWalk(&v)
	return v.Hist
}

// Gets the order of instantiation of the , by traversing the graph breadth-first
func (graph *serviceGraph) GetServiceOrderById(id string) col.IStack[string] {
	subGraph, _, _ := graph.dag.GetDescendantsGraph(id)
	v := depVisiter{Hist: col.NewStack[string](nil)}
	subGraph.BFSWalk(&v)
	return v.Hist
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
	if service == nil {
		return errors.New("Cannot add nil service")
	}
	id, err := graph.dag.AddVertex(service)
	service.Id = id
	if err != nil {
		return err
	}
	return nil
}

func (graph *serviceGraph) AddEdge(srcService *Service, depService *Service) error {
	if srcService == nil || depService == nil {
		return errors.New("Cannot add edge to or from nil")
	}
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
	pv.Hist.Push(id)
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
