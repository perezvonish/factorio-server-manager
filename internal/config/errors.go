package config

import (
	"errors"
	"fmt"
)

var (
	WhileLoadingFileError   = errors.New(`error on config file loading`)
	WhileLoadingFieldsError = errors.New(`error on config fields loading`)
)

type FieldRequiredError struct {
	envName string
}

func (f FieldRequiredError) Error() string {
	return fmt.Sprintf("required environment variable %s is not set", f.envName)
}

func NewFieldRequiredError(envName string) FieldRequiredError {
	return FieldRequiredError{envName}
}
