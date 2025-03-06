// Configuration loading from various sources
//
// Supports loading from:
// - YAML/JSON files
// - Environment variables
// - Command line flags
//
// Handles merging of configuration from multiple sources
// with proper precedence rules

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from the specified file path and
// overrides with environment variables.
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}
	
	// Set defaults first
	SetDefaults(config)
	
	// Try to load from file if provided
	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}
	
	// Override with environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load config from environment: %w", err)
	}
	
	return config, nil
}

// loadFromFile loads the configuration from a YAML file
func loadFromFile(config *Config, path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", path)
	}
	
	// Read file content
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	
	// Parse YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}
	
	return nil
}

// loadFromEnv overrides configuration with values from environment variables
func loadFromEnv(config *Config) error {
	prefix := "ILINDEN_"
	
	// Get all environment variables
	envVars := os.Environ()
	for _, env := range envVars {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		
		// Split key and value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := parts[0]
		value := parts[1]
		
		// Remove prefix and build path
		key = strings.TrimPrefix(key, prefix)
		path := strings.Split(strings.ToLower(key), "_")
		
		// Set the value in the config struct
		if err := setConfigValue(config, path, value); err != nil {
			return err
		}
	}
	
	return nil
}

// setConfigValue sets a value in the config struct at the specified path
func setConfigValue(config *Config, path []string, value string) error {
	if len(path) == 0 {
		return nil
	}
	
	// Start with the config object
	val := reflect.ValueOf(config).Elem()
	
	// Navigate through the config struct to the target field
	for i, part := range path {
		// Capitalize the first letter of the part to match Go's exported fields
		fieldName := strings.ToUpper(part[:1]) + part[1:]
		
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			return fmt.Errorf("config field not found: %s", strings.Join(path[:i+1], "."))
		}
		
		// If this is the final part of the path, set the value
		if i == len(path)-1 {
			return setFieldValue(field, value)
		}
		
		// Otherwise, keep traversing
		if field.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct for field %s, got %s", fieldName, field.Kind())
		}
		
		val = field
	}
	
	return nil
}

// setFieldValue sets a field's value based on its type
func setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}
	
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		field.SetBool(boolVal)
	
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special handling for time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value: %s", value)
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			field.SetInt(intVal)
		}
	
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatVal)
	
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// Parse comma-separated list
			items := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), len(items), len(items))
			for i, item := range items {
				slice.Index(i).SetString(strings.TrimSpace(item))
			}
			field.Set(slice)
		} else {
			return fmt.Errorf("unsupported slice type: %s", field.Type().Elem().Kind())
		}
	
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	
	return nil
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	// JWT validation if enabled
	if c.JWT.Enabled {
		if c.JWT.Secret == "" && c.JWT.KeysURL == "" {
			return fmt.Errorf("JWT is enabled but neither Secret nor KeysURL is provided")
		}
	}
	
	// Redis validation if enabled
	if c.Redis.Enabled && len(c.Redis.Addresses) == 0 {
		return fmt.Errorf("Redis is enabled but no addresses are provided")
	}
	
	return nil
}

// GetAddress returns the full server address with host and port
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}