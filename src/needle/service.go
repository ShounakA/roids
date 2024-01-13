/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package needle

import "reflect"

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
func AddService[T interface{}](spec T, impl any) {
	// Get IoC Container
	container := GetRoids()
	specType := reflect.TypeOf(spec).Elem()

	// Add vertex for the service being added
	thisV, _ := servicesGraph.AddVertex(specType)
	container.services[specType] = &Service{Id: thisV, Injector: impl}

	ftype := reflect.TypeOf(impl)
	// Get all dependencies in injector
	for i := 0; i < ftype.NumIn(); i++ {
		field := ftype.In(i)
		// Add vertex for dependency
		depV, err := servicesGraph.AddVertex(field)
		if err != nil {
			depV = container.services[field].Id
		}
		// Add edge
		err = servicesGraph.AddEdge(thisV, depV)
		if err != nil {
			println(err.Error())
		}
	}
}

// Gets an implementation of a service based on an specification from the container.
func Inject[T interface{}]() T {
	c := GetRoids()
	implType := reflect.TypeOf(new(T)).Elem()
	return (*(c.services[implType].instance)).(T)
}
