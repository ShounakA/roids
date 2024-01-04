/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package needle

import (
	"context"
	"fmt"
	"reflect"
	"roids/col"
	"sync"

	"github.com/heimdalr/dag"
)

// Needle is a struct that holds all the dependencies for the application.
// Though you can create this struct it is recommended to use the `GetNeedle` function to get the global instance.
type Needle struct {
	services map[reflect.Type]*Service
	context  context.Context
}

// Thread-safe function to get the global instance of the dependency container.
func GetNeedle() *Needle {
	once.Do(func() {
		instance = NewDependencyContainer()
	})
	return instance
}

// Creates a new instance of the dependency container.
// This function should not be used directly. Use `GetContainer` instead.
func NewDependencyContainer() *Needle {
	return &Needle{
		services: services,
	}
}

// Method to traverse the entire dependency graph
func (pv *depVisiter) Visit(v dag.Vertexer) {
	id, value := v.Vertex()
	sType := value.(reflect.Type)
	pv.Hist.Push(sType)
	isLeaf, err := servicesGraph.IsLeaf(id)
	if err != nil {
		println("error")
	}
	services[sType].isLeaf = isLeaf
}

// Build the dependency injection IoC container
func Build() {
	// Instantiate/Get IoC Container
	_ = GetNeedle()

	v := depVisiter{Hist: col.NewStack[reflect.Type](nil)}
	servicesGraph.BFSWalk(&v)

	for v.Hist.GetSize()-1 > 0 {
		serviceType := *v.Hist.Pop()
		service := services[serviceType]
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
				fmt.Println("Instance is not a function")
			}
		} else {
			fmt.Println("Could not create service.")
		}
	}
}

// Prints all dependencies in the container
func PrintDependencyGraph() {
	fmt.Println(servicesGraph.String())
}

/**
NOT EXPORTED STUFF
vvvvvvvvvvvvvvvvv
*/

// Application wide instance of the dependency container.
var instance *Needle

// Map of all the services in the container.
var services = make(map[reflect.Type]*Service)

// Atomic boolean to ensure that the container is only created once.
var once sync.Once

// Graph of dependent services.
var servicesGraph = dag.NewDAG()

// Dependency visitor. It keeps track of the nodes visited into a stack,
// so that we can instantiate leaf deps by popping them out.
type depVisiter struct {
	// History of the dependent services visited.
	Hist col.IStack[reflect.Type]
}

// Get all deps before using injector.
func getArgsForFunction(service *Service) []reflect.Value {
	injected := service.Injector
	injectedVal := reflect.ValueOf(injected)
	injectedType := injectedVal.Type()

	argValues := make([]reflect.Value, injectedType.NumIn())

	// Get the type of each argument
	for i := 0; i < injectedType.NumIn(); i++ {
		serviceType := injectedType.In(i)
		instanceVal := reflect.ValueOf(*(services[serviceType].instance))
		argValues[i] = instanceVal
	}
	return argValues
}

// Creates an instance of a leaf service.
// These services should not have parameters in there injector functions.
// Meaning they can be created easily without problem.
func createLeafDep(sType reflect.Type) {
	injected := services[sType].Injector
	injectedVal := reflect.ValueOf(injected)
	if injectedVal.Kind() == reflect.Func {
		// If the instance is a function, call it
		results := injectedVal.Call(nil)
		// Handle results if necessary
		instance := results[0].Interface()
		services[sType].instance = &instance
		services[sType].created = true

	} else {
		fmt.Println("Instance is not a function")
	}
}
