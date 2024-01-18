package errors

import (
	"errors"
	"fmt"
	"reflect"
)

func NewNeedleError(message string, sType reflect.Type) error {
	msg := fmt.Sprintf("%s -> %s", sType.Name(), message)
	return errors.New(msg)
}

// ServiceError
// InjectorError
// CircularDepError
// DuplicateEdgeError -> Better Name
// UnknownError
