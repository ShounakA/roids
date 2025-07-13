/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/ShounakA/roids/core"
)

// Thread-safe function to get the global instance of the dependency container.
func GetRoids() *roidsContainer {
	once.Do(func() {
		globalRoidsContainer = newRoidsContainer(nil)
	})
	return globalRoidsContainer
}

// Builds all static services in container.
func Build() error {
	roids := GetRoids()
	startTime := time.Now()
	roids.Logger.Debug("Building static services:")
	order := roids.servicesGraph.getInstantiationOrder()
	roids.Logger.Debug(order.String())
	for order.GetSize() > 0 {
		vertexId := *order.Pop()
		service, _ := roids.servicesGraph.getVertex(vertexId)
		roids.Logger.Debug(fmt.Sprintf("Building static service %s:%s", service.ID(), service.SpecType.String()))
		if service.lifetimeType == core.StaticLifetime {
			if service.isRoot && !service.created {
				roids.Logger.Debug("Creating leaf service...")
				setStaticLeafDep(service)
			} else if !service.isRoot && !service.created {
				roids.Logger.Debug("Creating branch service...")
				setStaticBranchDep(service)
			} else {
				return core.NewUnknownError(nil)
			}
		}
	}
	roids.Logger.Debug(fmt.Sprintf("Completed building all services in %dµs", time.Since(startTime).Microseconds()))
	return nil
}

// Clears the container of all services
// SUPER UNSAFE. Only used during testing. Dont use while running an application.
func UNSAFE_Clear() {
	roids := GetRoids()
	roids.servicesGraph.clearGraph()
}

/**
 * Non-exported stuff
 */

// Application wide globalRoidsContainer of the dependency container.
var globalRoidsContainer *roidsContainer

// Atomic boolean to ensure that the container is only created once.
var once sync.Once

// Build a new instance of the specified service.
func buildTransientDep(service *Service) *any {
	roids := GetRoids()
	roids.Logger.Debug(fmt.Sprintf("Building transient service %s:%s", service.ID(), service.SpecType.String()))
	hist := roids.servicesGraph.getServiceOrderById(service.Id)
	deps := roids.servicesGraph.staticDepsMap
	roids.Logger.Debug(hist.String())
	for hist.GetSize() > 0 {
		id := *hist.Pop()
		service, err := roids.servicesGraph.getVertex(id)
		if err != nil {
			log.Panicf("Should have the vertex in the graph")
		}
		roids.Logger.Debug(fmt.Sprintf("Fetching dependant service %s:%s", service.ID(), service.SpecType.String()))
		switch service.lifetimeType {
		case core.StaticLifetime:
			deps[service.SpecType] = service.instance
		case core.TransientLifetime:
			if service.isRoot {
				roids.Logger.Debug("Creating leaf service...")
				transService := createTransientLeafDep(service)
				deps[service.SpecType] = transService
			} else {
				roids.Logger.Debug("Creating branch service...")
				transService := createTransientBranchDep(service, deps)
				deps[service.SpecType] = transService
			}
		}
	}

	transientDep := deps[service.SpecType]
	return transientDep
}

// Get all deps before using injector.
func getArgsForFunction(service *Service) []reflect.Value {
	roids := GetRoids()
	roids.Logger.Debug("Injecting services from injector function")
	injected := service.Injector
	injectedVal := reflect.ValueOf(injected)
	injectedType := injectedVal.Type()

	argValues := make([]reflect.Value, injectedType.NumIn())

	// Get the type of each argument
	for i := 0; i < injectedType.NumIn(); i++ {
		serviceType := injectedType.In(i)
		service := roids.servicesGraph.getServiceByType(serviceType)
		if service.lifetimeType == core.StaticLifetime {
			roids.Logger.Debug(fmt.Sprintf("Injecting static service %s:%s", service.ID(), service.SpecType.String()))
			instanceVal := reflect.ValueOf(*(service.instance))
			argValues[i] = instanceVal
		} else {
			roids.Logger.Debug(fmt.Sprintf("Injecting transient service %s:%s", service.ID(), service.SpecType.String()))
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
	r := GetRoids()
	service.instance = createTransientLeafDep(service)
	r.servicesGraph.staticDepsMap[service.SpecType] = service.instance
	service.created = true
}

// Sets a static instance of a branch or root dependency.
// Static services can depend on Transient services,
// so we may need to create build one
func setStaticBranchDep(service *Service) {
	r := GetRoids()
	injector := service.Injector
	injectorVal := reflect.ValueOf(injector)
	args := getArgsForFunction(service)
	results := injectorVal.Call(args)
	newStaticService := results[0].Interface()
	service.instance = &newStaticService
	r.servicesGraph.staticDepsMap[service.SpecType] = service.instance
	service.created = true
	return
}

// roidsContainer is a struct that holds all the dependencies for the application.
// It is recommended to use the `GetRoids` function to get the global instance.
type roidsContainer struct {
	servicesGraph *serviceGraph
	Logger        *slog.Logger
}

// Creates a new instance of the dependency container.
// This function should not be used directly. Use `GetRoids` instead.
func newRoidsContainer(graph *serviceGraph) *roidsContainer {
	logFile, _ := os.Create("roids.log")
	libLogger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	if graph == nil {
		dag2 := core.NewGraph()
		return &roidsContainer{
			servicesGraph: newServiceGraph(dag2),
			Logger:        libLogger,
		}
	} else {
		return &roidsContainer{
			servicesGraph: graph,
			Logger:        libLogger,
		}
	}
}
