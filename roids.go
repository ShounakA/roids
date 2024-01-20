/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/ShounakA/roids/col"

	"github.com/heimdalr/dag"
)

// Thread-safe function to get the global instance of the dependency container.
func GetRoids() *needleContainer {
	once.Do(func() {
		instance = newNeedle()
	})
	return instance
}

// Method to traverse the entire dependency graph
func (pv *depVisiter) Visit(v dag.Vertexer) {
	roids := GetRoids()
	id, value := v.Vertex()
	service := value.(*Service)
	pv.Hist.Push(service.SpecType)
	isLeaf, _ := roids.servicesGraph.IsLeaf(id)
	roids.services[service.SpecType].isLeaf = isLeaf
}

// Build the dependency injection IoC container
func Build() error {
	// Instantiate/Get IoC Container
	roids := GetRoids()

	v := depVisiter{Hist: col.NewStack[reflect.Type](nil)}
	roids.servicesGraph.BFSWalk(&v)

	for v.Hist.GetSize() > 0 {
		serviceType := *v.Hist.Pop()
		service := roids.services[serviceType]
		if service.isLeaf && !service.created {
			createLeafDep(serviceType)
		} else if !service.isLeaf && !service.created {
			injected := service.Injector
			injectedVal := reflect.ValueOf(injected)
			if injectedVal.Kind() == reflect.Func {
				args := getArgsForFunction(service)
				results := injectedVal.Call(args)
				instance := results[0].Interface()
				service.instance = &instance
				service.created = true

			} else {
				return NewInjectorError(serviceType)
			}
		} else if service.created {
			fmt.Println("Could not create service. It is already created.")
		} else {
			return NewUnknownError(nil)
		}
	}
	return nil
}

// Prints all dependencies in the container
func PrintDependencyGraph() {
	roids := GetRoids()
	fmt.Println(roids.servicesGraph.String())
}

// Clears the container of all services
// SUPER UNSAFE. Only used during testing. Dont use while running an application.
func UNSAFE_Clear() {
	roids := GetRoids()
	for t := range roids.services {
		delete(roids.services, t)
	}
	roids.servicesGraph = dag.NewDAG()
}

/**
Private stuff
*/

// Application wide instance of the dependency container.
var instance *needleContainer

// Atomic boolean to ensure that the container is only created once.
var once sync.Once

// Dependency visitor. It keeps track of the nodes visited into a stack,
// so that we can instantiate leaf deps by popping them out.
type depVisiter struct {
	// History of the dependent services visited.
	Hist col.IStack[reflect.Type]
}

// Get all deps before using injector.
func getArgsForFunction(service *Service) []reflect.Value {
	roids := GetRoids()
	injected := service.Injector
	injectedVal := reflect.ValueOf(injected)
	injectedType := injectedVal.Type()

	argValues := make([]reflect.Value, injectedType.NumIn())

	// Get the type of each argument
	for i := 0; i < injectedType.NumIn(); i++ {
		serviceType := injectedType.In(i)
		instanceVal := reflect.ValueOf(*(roids.services[serviceType].instance))
		argValues[i] = instanceVal
	}
	return argValues
}

// Creates an instance of a leaf service.
// These services should not have parameters in there injector functions.
// Meaning they can be created easily without problem.
func createLeafDep(sType reflect.Type) error {
	roids := GetRoids()
	injected := roids.services[sType].Injector
	injectedVal := reflect.ValueOf(injected)
	if injectedVal.Kind() == reflect.Func {
		// If the instance is a function, call it
		results := injectedVal.Call(nil)
		// Handle results if necessary
		instance := results[0].Interface()
		roids.services[sType].instance = &instance
		roids.services[sType].created = true

	} else {
		return NewInjectorError(sType)
	}
	return nil
}

// needleContainer is a struct that holds all the dependencies for the application.
// It is recommended to use the `GetNeedle` function to get the global instance.
type needleContainer struct {
	services      map[reflect.Type]*Service
	servicesGraph *dag.DAG
	context       context.Context
}

// Creates a new instance of the dependency container.
// This function should not be used directly. Use `GetNeedle` instead.
func newNeedle() *needleContainer {
	return &needleContainer{
		services:      make(map[reflect.Type]*Service),
		servicesGraph: dag.NewDAG(),
	}
}
