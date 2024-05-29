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
		dag *core.AcyclicGraph
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
		dag: d2,
	}
}

// Gets the order of instantiation, by traversing the graph breadth-first
func (graph *serviceGraph) getInstantiationOrder() col.IStack[string] {
	v := depVisiter{Hist: col.NewStack[string](nil)}
	graph.dag.TraverseBF(&v)
	return v.Hist
}

// Gets the order of instantiation of the , by traversing the graph breadth-first
func (graph *serviceGraph) getServiceOrderById(id string) col.IStack[string] {
	v := depVisiter{Hist: col.NewStack[string](nil)}
	graph.dag.TraverseBFFrom(id, &v)
	return v.Hist
}

// Gets the Service struct from the graph by the interface type provided.
func (graph *serviceGraph) getServiceByType(specType reflect.Type) *Service {
	tmpService := Service{SpecType: specType}
	if node, err := graph.dag.GetVertex(tmpService.ID()); err != nil {
		return nil
	} else {
		return node.Value().(*Service)
	}
}

func (graph *serviceGraph) getVertex(id string) (*Service, error) {
	vertex, err := graph.dag.GetVertex(id)
	if err != nil {
		return nil, err
	}
	service := vertex.Value().(*Service)
	return service, nil
}

func (graph *serviceGraph) addVertex(service *Service) error {
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

func (graph *serviceGraph) addEdge(srcService *Service, depService *Service) error {
	if srcService == nil || depService == nil {
		return errors.New("Cannot add edge to or from nil")
	}
	err := graph.dag.AddEdge(srcService.Id, depService.Id)
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
func (graph *serviceGraph) clearGraph() {
	graph.dag = core.NewGraph()
}

func (pv *depVisiter) Do(v *core.Traverser) {
	service := v.GetVertex().Value().(*Service)
	pv.Hist.Push(service.Id)
	service.isLeaf = v.GetVertex().IsLeaf()
}
