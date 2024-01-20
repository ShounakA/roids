/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"reflect"

	"github.com/heimdalr/dag"
)

// Struct representing an injectable service. (aka Provider, Assembler, Service, or Injector)
type Service struct {
	Injector any
	Id       string
	created  bool
	implType reflect.Type
	instance *any
	isLeaf   bool
}

// Adds a service to the container.
// Uses the specification (interface or struct) to inject an implmentation into the IoC container
func AddService[T interface{}](spec T, impl any) error {

	// Get IoC Container
	container := GetRoids()
	specType := reflect.TypeOf(spec).Elem()
	if reflect.ValueOf(impl).Kind() != reflect.Func {
		return NewInjectorError(specType)
	}
	ftype := reflect.TypeOf(impl)
	implType := ftype.Out(0)
	if !implType.Implements(specType) {
		return NewServiceError(specType, implType.Elem())
	}

	// Add vertex for the service being added
	thisV, err := container.servicesGraph.AddVertex(specType)
	if err != nil {
		// It means we added a vertex for this service before via a constructor.
		// SO we must lookup the id based on the service type.
		lookup := &reverseLookupVisiter{searchType: specType}
		container.servicesGraph.BFSWalk(lookup)
		thisV = lookup.vertexId
	}
	container.services[specType] = &Service{Id: thisV, Injector: impl}

	// Get all dependencies in injector
	for i := 0; i < ftype.NumIn(); i++ {
		field := ftype.In(i)
		// Add vertex for dependency
		depV, err := container.servicesGraph.AddVertex(field)
		if err != nil {
			depV = container.services[field].Id
		}
		// Add edge
		err = container.servicesGraph.AddEdge(thisV, depV)
		if err != nil {
			switch e := err.(type) {
			case dag.EdgeLoopError:
				return NewCircularDependencyError(e, specType)
			case dag.EdgeDuplicateError:
				return NewDuplicateEdgeError(e, thisV, specType)
			default:
				return NewUnknownError(e)
			}
		}
	}
	return nil
}

// Gets an implementation of a service based on an specification from the container.
func Inject[T interface{}]() T {
	c := GetRoids()
	implType := reflect.TypeOf(new(T)).Elem()
	return (*(c.services[implType].instance)).(T)
}

type reverseLookupVisiter struct {
	vertexId   string
	searchType reflect.Type
}

// Function to lookup vertexId based on spec
func (pv *reverseLookupVisiter) Visit(v dag.Vertexer) {
	id, value := v.Vertex()
	sType := value.(reflect.Type)
	if sType == pv.searchType {
		pv.vertexId = id
		return
	}
}
