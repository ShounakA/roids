/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"errors"
	"reflect"

	"github.com/ShounakA/roids/col"
	"github.com/ShounakA/roids/core"
)

type (
	serviceGraph struct {
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
func newServiceGraph(d2 *core.AcyclicGraph) *serviceGraph {
	return &serviceGraph{
		dag2: d2,
	}
}

// Gets the order of instantiation, by traversing the graph breadth-first
func (graph *serviceGraph) GetInstantiationOrder() col.IStack[string] {
	v := depVisiter{Hist: col.NewStack[string](nil)}
	graph.dag2.TraverseBF(&v)
	return v.Hist
}

// Gets the order of instantiation of the , by traversing the graph breadth-first
func (graph *serviceGraph) GetServiceOrderById(id string) col.IStack[string] {
	v := depVisiter{Hist: col.NewStack[string](nil)}
	graph.dag2.TraverseBFFrom(id, &v)
	return v.Hist
}

// Gets the Service struct from the graph by the interface type provided.
func (graph *serviceGraph) GetServiceByType(specType reflect.Type) *Service {
	tmpService := Service{SpecType: specType}
	if node, err := graph.dag2.GetVertex(tmpService.ID()); err != nil {
		return nil
	} else {
		return node.Value().(*Service)
	}
}

func (graph *serviceGraph) GetVertex(id string) (*Service, error) {
	vertex, err := graph.dag2.GetVertex(id)
	if err != nil {
		return nil, err
	}
	service := vertex.Value().(*Service)
	return service, nil
}

func (graph *serviceGraph) AddVertex(service *Service) error {
	if service == nil {
		return errors.New("Cannot add nil service")
	}
	id, err := graph.dag2.AddVertex(service)
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
	err := graph.dag2.AddEdge(srcService.Id, depService.Id)
	if err != nil {
		switch e := err.(type) {
		case *core.EdgeCycleError:
			return core.NewCircularDependencyError(e, srcService.SpecType)
		case *core.EdgeExistsError:
			return core.NewDuplicateEdgeError(e, srcService.Id, srcService.SpecType)
		default:
			return core.NewUnknownError(e)
		}
	}
	return nil
}

// Function to clear the services graph
func (graph *serviceGraph) ClearGraph() {
	graph.dag2 = core.NewGraph()
}

// TODO wrap core.Node with a different struct so its not accessible
func (pv *depVisiter) Do(v *core.Node) {
	service := v.Value().(*Service)
	pv.Hist.Push(service.Id)
	service.isLeaf = v.IsLeaf()
}
