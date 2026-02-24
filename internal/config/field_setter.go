package config

import (
	"fmt"
	"reflect"
	"strconv"
)

func setField(field reflect.Value, value string, envName string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			return nil
		}
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int value for %s: %v", envName, err)
		}
		field.SetInt(intVal)

	default:
		return fmt.Errorf("unsupported type %s for field %s", field.Kind(), envName)
	}

	return nil
}
