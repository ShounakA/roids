package roids

import (
	"fmt"
	"reflect"
)

type (
	ServiceError struct {
		SpecType reflect.Type
		ImplType reflect.Type
	}

	InjectorError struct {
		err      error
		SpecType reflect.Type
	}

	CircularDependencyError struct {
		err      error
		SpecType reflect.Type
	}

	DuplicateEdgeError struct {
		err      error
		VertexId string
		SpecType reflect.Type
	}

	InvalidLifetimeError struct {
		err      error
		SpecType reflect.Type
	}

	UnknownError struct {
		err error
	}
)

func NewServiceError(spec reflect.Type, impl reflect.Type) *ServiceError {
	return &ServiceError{
		SpecType: spec,
		ImplType: impl,
	}
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("[%s] '%s' must implement '%s' to be added as a service.", e.SpecType, e.SpecType, e.ImplType)
}

func NewInjectorError(spec reflect.Type) *InjectorError {
	return &InjectorError{
		SpecType: spec,
	}
}

func (e *InjectorError) Error() string {
	return fmt.Sprintf("[%s] Injector is not a function. -> %s", e.SpecType, e.err.Error())
}

func NewCircularDependencyError(err error, spec reflect.Type) *CircularDependencyError {
	return &CircularDependencyError{
		err:      err,
		SpecType: spec,
	}
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("[%s] Circular dependency detected. -> %s", e.SpecType, e.err.Error())
}

func NewDuplicateEdgeError(err error, id string, spec reflect.Type) *DuplicateEdgeError {
	return &DuplicateEdgeError{
		err:      err,
		VertexId: id,
		SpecType: spec,
	}
}

func (e *DuplicateEdgeError) Error() string {
	return fmt.Sprintf("[%s] Duplicate service and dependency detected. -> %s", e.SpecType, e.err.Error())
}

func NewInvalidLifetimeError(err error, spec reflect.Type) *InvalidLifetimeError {
	return &InvalidLifetimeError{
		err:      err,
		SpecType: spec,
	}
}

func (e *InvalidLifetimeError) Error() string {
	return fmt.Sprintf("[%s] Invalid lifetime. Valid  lifetimes are: %s and %s", e.SpecType, StaticService, TransientService)
}

func NewUnknownError(err error) *UnknownError {
	return &UnknownError{
		err: err,
	}
}

func (e *UnknownError) Error() string {
	return fmt.Sprintf("Unknown error occurred. -> %s", e.err.Error())
}
