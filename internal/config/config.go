package config

import (
	"os"
	"reflect"
)

func Init() (*Config, error) {
	config, err := load()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func load() (*Config, error) {
	if err := loadEnvFile(".env"); err != nil {
		if !os.IsNotExist(err) {
			return nil, WhileLoadingFileError
		}
	}

	config := &Config{}

	if err := loadFromEnv(config); err != nil {
		return nil, err
	}

	return config, nil
}

func processStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := processStruct(field); err != nil {
				return err
			}
			continue
		}

		envTag := fieldType.Tag.Get("env")
		defaultTag := fieldType.Tag.Get("envDefault")
		requiredTag := fieldType.Tag.Get(string(Required))

		if envTag == "" {
			continue
		}

		envVal := os.Getenv(envTag)

		if err := validateRequired(envTag, requiredTag, envVal); err != nil {
			return err
		}

		if envVal == "" {
			envVal = defaultTag
		}

		if err := setField(field, envVal, envTag); err != nil {
			return err
		}
	}

	return nil
}
