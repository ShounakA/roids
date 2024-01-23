/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"reflect"
	"sync"

	"github.com/heimdalr/dag"
)

// Thread-safe function to get the global instance of the dependency container.
func GetRoids() *roidsContainer {
	once.Do(func() {
		instance = newRoidsContainer(nil)
	})
	return instance
}

// Builds all static services in container.
func Build() error {
	roids := GetRoids()

	order := roids.servicesGraph.GetInstantiationOrder()

	for order.GetSize() > 0 {
		serviceType := *order.Pop()
		service := roids.servicesGraph.GetServiceByType(serviceType)
		if service.lifetimeType == StaticLifetime {
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

// Prints all dependencies in the container
func PrintDependencyGraph() {
	roids := GetRoids()
	roids.servicesGraph.ShowGraph()
}

// Clears the container of all services
// SUPER UNSAFE. Only used during testing. Dont use while running an application.
func UNSAFE_Clear() {
	roids := GetRoids()
	roids.servicesGraph.ClearGraph()
}

/**
 * Non-exported stuff
 */

// Application wide instance of the dependency container.
var instance *roidsContainer

// Atomic boolean to ensure that the container is only created once.
var once sync.Once

// Build a new instance of the specified service.
func buildTransientDep(service *Service) *any {
	roids := GetRoids()
	hist := roids.servicesGraph.GetServiceOrderById(service.Id)
	deps := make(map[reflect.Type]*any)

	for hist.GetSize() > 0 {
		id := *hist.Pop()
		service, _ := roids.servicesGraph.GetVertex(id)
		if service.lifetimeType == StaticLifetime {
			deps[service.SpecType] = service.instance
		} else if service.lifetimeType == TransientLifetime {
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
		service := roids.servicesGraph.GetServiceByType(serviceType)
		if service.lifetimeType == StaticLifetime {
			instanceVal := reflect.ValueOf(*(service.instance))
			argValues[i] = instanceVal
		} else {
			dep := buildTransientDep(service)
			instanceVal := reflect.ValueOf(*dep)
			argValues[i] = instanceVal
		}
	}
	return argValues
}

// Creates a new leaf instance of the specified service
func createTransientLeafDep(service *Service) *any {
	injector := service.Injector
	injectorVal := reflect.ValueOf(injector)
	results := injectorVal.Call(nil)
	leafDep := results[0].Interface()
	return &leafDep
}

// Creates a new branch or root instance of the specified service
func createTransientBranchDep(service *Service, deps map[reflect.Type]*any) *any {
	injectorVal := reflect.ValueOf(service.Injector)
	injectorType := injectorVal.Type()

	argValues := make([]reflect.Value, injectorType.NumIn())
	for i := 0; i < injectorType.NumIn(); i++ {
		serviceType := injectorType.In(i)
		dep := deps[serviceType]
		instanceVal := reflect.ValueOf(*dep)
		argValues[i] = instanceVal
	}
	results := injectorVal.Call(argValues)
	dep := results[0].Interface()
	return &dep
}

// Sets a static instance of a leaf service.
// These services should not have parameters in there injector functions.
// Meaning they can be created by calling the injector.
func setStaticLeafDep(service *Service) {
	service.instance = createTransientLeafDep(service)
	service.created = true
}

// Sets a static instance of a branch or root dependency.
// Static services can depend on Transient services,
// so we may need to create build one
func setStaticBranchDep(service *Service) {
	injector := service.Injector
	injectorVal := reflect.ValueOf(injector)
	args := getArgsForFunction(service)
	results := injectorVal.Call(args)
	instance := results[0].Interface()
	service.instance = &instance
	service.created = true
	return
}

// roidsContainer is a struct that holds all the dependencies for the application.
// It is recommended to use the `GetNeedle` function to get the global instance.
type roidsContainer struct {
	servicesGraph *serviceGraph
}

// Creates a new instance of the dependency container.
// This function should not be used directly. Use `GetNeedle` instead.
func newRoidsContainer(graph *serviceGraph) *roidsContainer {
	if graph == nil {
		dag := dag.NewDAG()
		return &roidsContainer{
			servicesGraph: newServiceGraph(dag),
		}
	} else {
		return &roidsContainer{
			servicesGraph: graph,
		}
	}
}
