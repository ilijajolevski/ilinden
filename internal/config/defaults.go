// Default configuration values
package config

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SetDefaults populates default values for configuration fields
// based on the 'default' struct tags
func SetDefaults(config *Config) {
	setDefaultsForStruct(reflect.ValueOf(config).Elem())
}

// setDefaultsForStruct recursively sets default values for all fields in a struct
func setDefaultsForStruct(val reflect.Value) {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)
		
		// Skip unexported fields
		if !field.CanSet() {
			continue
		}
		
		// Get default value from tag
		defaultValue := typeField.Tag.Get("default")
		if defaultValue == "" {
			// If no default value but field is a struct, process it recursively
			if field.Kind() == reflect.Struct {
				setDefaultsForStruct(field)
			}
			continue
		}
		
		// Set default value based on field type
		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				field.SetString(defaultValue)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() == 0 {
				// Handle duration specially
				if typeField.Type == reflect.TypeOf(time.Duration(0)) {
					duration, err := time.ParseDuration(defaultValue)
					if err == nil {
						field.Set(reflect.ValueOf(duration))
					}
				} else {
					intVal, err := strconv.ParseInt(defaultValue, 10, 64)
					if err == nil {
						field.SetInt(intVal)
					}
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() == 0 {
				uintVal, err := strconv.ParseUint(defaultValue, 10, 64)
				if err == nil {
					field.SetUint(uintVal)
				}
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() == 0 {
				floatVal, err := strconv.ParseFloat(defaultValue, 64)
				if err == nil {
					field.SetFloat(floatVal)
				}
			}
		case reflect.Bool:
			if !field.Bool() { // Default for false only
				boolVal, err := strconv.ParseBool(defaultValue)
				if err == nil {
					field.SetBool(boolVal)
				}
			}
		case reflect.Slice:
			if field.Len() == 0 {
				// Process array default value in format [\"value1\", \"value2\"]
				if field.Type().Elem().Kind() == reflect.String {
					trimmed := strings.Trim(defaultValue, "[]")
					if trimmed != "" {
						items := strings.Split(trimmed, ",")
						slice := reflect.MakeSlice(field.Type(), len(items), len(items))
						for i, item := range items {
							cleanItem := strings.Trim(strings.TrimSpace(item), "\"")
							slice.Index(i).SetString(cleanItem)
						}
						field.Set(slice)
					}
				}
			}
		case reflect.Struct:
			setDefaultsForStruct(field)
		}
	}
}