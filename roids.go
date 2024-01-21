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
func GetRoids() *roidsContainer {
	once.Do(func() {
		instance = newRoidsContainer()
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
	service.isLeaf = isLeaf
}

// Builds all static services in container.
func Build() error {
	roids := GetRoids()

	v := depVisiter{Hist: col.NewStack[reflect.Type](nil)}
	roids.servicesGraph.BFSWalk(&v)

	for v.Hist.GetSize() > 0 {
		serviceType := *v.Hist.Pop()
		service := GetServiceByType(roids.servicesGraph, serviceType)
		if service.lifetimeType == StaticService {
			if service.isLeaf && !service.created {
				setStaticLeafDep(service)
			} else if !service.isLeaf && !service.created {
				setStaticBranchDep(service)
			} else {
				return NewUnknownError(nil)
			}
		}
	}
	return nil
}

func BuildTransientDep(service *Service) *any {
	roids := GetRoids()
	chVertex, _, _ := roids.servicesGraph.DescendantsWalker(service.Id)
	hist := col.NewStack[string](&service.Id)
	select {
	case vertexId := <-chVertex:
		if vertexId != "" {
			hist.Push(vertexId)
		}
	}

	deps := make(map[reflect.Type]*any)

	for hist.GetSize() > 0 {
		id := *hist.Pop()
		vertex, _ := roids.servicesGraph.GetVertex(id)
		service := vertex.(*Service)
		if service.lifetimeType == StaticService {
			deps[service.SpecType] = service.instance
		} else if service.lifetimeType == TransientService {
			if service.isLeaf {
				instance := createTransientLeafDep(service)
				deps[service.SpecType] = instance
			} else {
				deps[service.SpecType] = createTransientBranchDep(service, deps)
			}
		}
	}

	transientDep := deps[service.SpecType]
	return transientDep
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
 * Private stuff
 */

// Application wide instance of the dependency container.
var instance *roidsContainer

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
		service := GetServiceByType(roids.servicesGraph, serviceType)
		if service.lifetimeType == StaticService {
			instanceVal := reflect.ValueOf(*(service.instance))
			argValues[i] = instanceVal
		} else {
			dep := BuildTransientDep(service)
			instanceVal := reflect.ValueOf(*dep)
			argValues[i] = instanceVal
		}
	}
	return argValues
}

func createTransientLeafDep(service *Service) *any {
	injected := service.Injector
	injectedVal := reflect.ValueOf(injected)
	results := injectedVal.Call(nil)
	// Handle results if necessary
	leafDep := results[0].Interface()
	return &leafDep
}

func createTransientBranchDep(service *Service, deps map[reflect.Type]*any) *any {
	injectedVal := reflect.ValueOf(service.Injector)
	injectedType := injectedVal.Type()

	argValues := make([]reflect.Value, injectedType.NumIn())
	for i := 0; i < injectedType.NumIn(); i++ {
		serviceType := injectedType.In(i)
		dep := deps[serviceType]
		instanceVal := reflect.ValueOf(*dep)
		argValues[i] = instanceVal
	}
	results := injectedVal.Call(argValues)
	dep := results[0].Interface()
	return &dep
}

// Creates an instance of a leaf service.
// These services should not have parameters in there injector functions.
// Meaning they can be created easily without problem.
func setStaticLeafDep(service *Service) {
	injected := service.Injector
	injectedVal := reflect.ValueOf(injected)
	results := injectedVal.Call(nil)
	// Handle results if necessary
	instance := results[0].Interface()
	service.instance = &instance
	service.created = true
}

func setStaticBranchDep(service *Service) {
	injected := service.Injector
	injectedVal := reflect.ValueOf(injected)
	args := getArgsForFunction(service)
	results := injectedVal.Call(args)
	instance := results[0].Interface()
	service.instance = &instance
	service.created = true
	return
}

// roidsContainer is a struct that holds all the dependencies for the application.
// It is recommended to use the `GetNeedle` function to get the global instance.
type roidsContainer struct {
	services      map[reflect.Type]*Service
	servicesGraph *dag.DAG
	context       context.Context
}

// Creates a new instance of the dependency container.
// This function should not be used directly. Use `GetNeedle` instead.
func newRoidsContainer() *roidsContainer {
	return &roidsContainer{
		services:      make(map[reflect.Type]*Service),
		servicesGraph: dag.NewDAG(),
	}
}
