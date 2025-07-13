/**
 * Author: Shounak Amladi
 * Date Created: 25/12/2023
 */

// Package containing custom dependency container for dependency injection.
// There is only ever one container and it can be used globally to access all the dependencies.
package roids

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/ShounakA/roids/core"
	"github.com/ShounakA/roids/core/config"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Struct representing an injectable service. (aka Provider, Assembler, Service, or Injector)
type Service struct {
	// Injector is a function that returns a pointer to a concrete implementation
	Injector any
	// ID representing and identifying the service
	Id string
	// The service specification or interface type
	SpecType reflect.Type
	// The lifetime of the service. Can either be "static" or "transient"
	lifetimeType string
	// True if the service has already been created once. False otherwise.
	created bool
	// The service concrete implementation type
	implType reflect.Type
	// The instantiated service. nil for service with "transient" lifetime
	instance *any
	// True the dependency does not require another to be instantiated.
	isLeaf bool

	isRoot bool
}

// String function for *Service type.
func (s *Service) String() string {
	return fmt.Sprintf("%s:%s", s.lifetimeType, s.SpecType)
}

// ID function for *Service type.
func (s *Service) ID() string {
	return uuid.NewSHA1(uuid.UUID{}, []byte(s.SpecType.Name())).String()
}

// Adds a static service to the container. A static service is only created once and lives for the life of the application.
// Uses the specification (interface or struct) to inject an implementation into the IoC container
func AddStaticService[T interface{}](spec T, impl any) error {
	return addService(spec, impl, core.StaticLifetime)
}

// Adds a transient service to the container. A transient service is newly instantiated for each use.
func AddTransientService[T interface{}](spec T, impl any) error {
	return addService(spec, impl, core.TransientLifetime)
}

// Gets an implementation of a service based on an specification from the container.
func Inject[T interface{}]() T {
	c := GetRoids()

	// service definition
	specType := reflect.TypeOf(new(T)).Elem()

	// Implementation of service
	service := c.servicesGraph.getServiceByType(specType)
	var impl T
	if service.lifetimeType == core.StaticLifetime {
		impl = (*(service.instance)).(T)
		return impl
	}
	dep := buildTransientDep(service)
	impl = (*dep).(T)
	return impl
}

// Generic add service definition function.
func addService[T interface{}](spec T, impl any, lifeTime string) error {

	// Get Container
	container := GetRoids()

	// Check for argument errors
	specType := reflect.TypeOf(spec).Elem()
	if lifeTime != core.StaticLifetime && lifeTime != core.TransientLifetime {
		return core.NewInvalidLifetimeError(nil, specType)
	}

	if reflect.ValueOf(impl).Kind() != reflect.Func {
		return core.NewInjectorError(specType)
	}

	ftype := reflect.TypeOf(impl)
	implType := ftype.Out(0)
	if !implType.Implements(specType) {
		return core.NewServiceError(specType, implType.Elem())
	}

	// Add vertex for the service being added
	srcService := &Service{Injector: impl, lifetimeType: lifeTime, SpecType: specType}
	err := container.servicesGraph.addVertex(srcService)
	if err != nil {
		// It means we added a vertex for this service before via a constructor.
		// SO we must lookup the id based on the service type.
		service := container.servicesGraph.getServiceByType(specType)
		service.implType = implType
		service.Injector = impl
		service.lifetimeType = lifeTime
		service.SpecType = specType
		srcService = service
	}

	// Get all dependencies in injector
	for i := 0; i < ftype.NumIn(); i++ {
		field := ftype.In(i)
		// Add vertex for dependency
		depService := container.servicesGraph.getServiceByType(field)
		if depService == nil {
			// Ignore the error as service = nil meaning we should not get an error adding vertex.
			depService = &Service{SpecType: field}
			_ = container.servicesGraph.addVertex(depService)
			err = container.servicesGraph.addEdge(srcService, depService)
		} else {
			err = container.servicesGraph.addEdge(srcService, depService)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Add a custom configuration file.
// Default roids.settings.json file.
func AddConfigurationBuilder[T any](filePath string, cfgType core.ConfigType) error {
	// Read file
	readFile := func() ([]byte, error) {
		return os.ReadFile(filePath)
	}
	return AddCustomConfiguration[T](readFile, cfgType)
}

// Add Custom Configuration
// configuraitonRead expects a slice of bytes encoded as json or yaml depending on the config type.
func AddCustomConfiguration[T any](configurationRead func() ([]byte, error), cfgType core.ConfigType) error {
	setFile, err := configurationRead()
	if err != nil {
		return err
	}
	var configFile config.RoidsConfiguration[T]
	switch cfgType {
	case core.JsonConfig:
		err = json.Unmarshal(setFile, &configFile)
		if err != nil {
			return err
		}
	case core.YamlConfig:
		err = yaml.Unmarshal(setFile, &configFile)
		if err != nil {
			return err
		}
	}

	// add the service
	err = addService(config.Create[config.IConfiguration[T]](), func() *config.RoidsConfiguration[T] {
		return &configFile
	}, core.StaticLifetime)

	if err != nil {
		return err
	}
	return nil
}
