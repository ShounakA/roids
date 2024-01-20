/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"fmt"
	"hash"
	"reflect"

	"github.com/google/uuid"
	"github.com/heimdalr/dag"
)

// Struct representing an injectable service. (aka Provider, Assembler, Service, or Injector)
type Service struct {
	Injector     any
	Id           string
	SpecType     reflect.Type
	lifetimeType string
	created      bool
	implType     reflect.Type
	instance     *any
	isLeaf       bool
}

func (s *Service) String() string {
	return fmt.Sprintf("%s:%s", s.lifetimeType, s.SpecType)
}

func (s *Service) ID() string {
	return uuid.NewSHA1(uuid.UUID{}, []byte(s.String())).String()
}

// Adds a lifetime service to the container.
// Uses the specification (interface or struct) to inject an implmentation into the IoC container
func AddLifetimeService[T interface{}](spec T, impl any) error {

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
	lifeService := &Service{Injector: impl, lifetimeType: "Lifetime", SpecType: specType}
	container.services[specType] = lifeService
	thisV, err := container.servicesGraph.AddVertex(lifeService)
	lifeService.Id = thisV
	if err != nil {
		// It means we added a vertex for this service before via a constructor.
		// SO we must lookup the id based on the service type.
		service := GetServiceByType(container.servicesGraph, specType)
		service.implType = implType
		service.Injector = impl
		service.lifetimeType = lifeService.lifetimeType
		service.SpecType = lifeService.SpecType
		service.Id = lifeService.Id
		thisV = lifeService.Id
	}

	// Get all dependencies in injector
	for i := 0; i < ftype.NumIn(); i++ {
		field := ftype.In(i)
		// Add vertex for dependency
		depService := GetServiceByType(container.servicesGraph, field)
		if depService == nil {
			// Ignore the error as service = nil meaning we should not get an error adding vertex.
			depV, _ := container.servicesGraph.AddVertex(&Service{SpecType: field, lifetimeType: "Lifetime"})
			err = container.servicesGraph.AddEdge(thisV, depV)
		} else {
			err = container.servicesGraph.AddEdge(thisV, depService.Id)
		}
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

func GetServiceByType(graph *dag.DAG, specType reflect.Type) *Service {
	lookup := &reverseLookupVisiter{searchType: specType, Service: nil}
	graph.BFSWalk(lookup)
	return lookup.Service
}

func AddTransientService[T interface{}](spec T, impl any) error {
	container := GetRoids()
	err := AddLifetimeService(spec, impl)
	if err != nil {
		return err
	}
	container.services[reflect.TypeOf(spec).Elem()].lifetimeType = "Transient"
	return nil
}

// Gets an implementation of a service based on an specification from the container.
func Inject[T interface{}]() T {
	c := GetRoids()
	specType := reflect.TypeOf(new(T)).Elem()

	// Implementation of service
	impl := (*(c.services[specType].instance)).(T)
	return impl
}

type reverseLookupVisiter struct {
	vertexId   string
	Service    *Service
	searchType reflect.Type
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

var serviceHash *hash.Hash = new(hash.Hash)
