package schemavalidator

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// readYAMLFile reads the YAML file and unmarshals it into a generic map
func readYAMLFile(filePath string) (map[string]interface{}, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(fileContent, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// resolveRefs recursively resolves $ref fields in the OpenAPI schema
func resolveRefs(openAPISpec map[string]interface{}, basePath string) (map[string]interface{}, error) {
	components, ok := openAPISpec["components"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("components section not found in OpenAPI spec")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schemas section not found in OpenAPI spec")
	}

	return resolveSchemaRefs(schemas, schemas, basePath)
}

// resolveSchemaRefs recursively resolves $ref within schemas and adds additionalProperties: false
func resolveSchemaRefs(schemas map[string]interface{}, currentSchema map[string]interface{}, basePath string) (map[string]interface{}, error) {
	for key, value := range currentSchema {
		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively resolve nested schemas
			_, err := resolveSchemaRefs(schemas, v, basePath)
			if err != nil {
				return nil, err
			}
		case string:
			// Resolve $ref fields
			if key == "$ref" {
				// Resolve reference path (assumes internal references using '#/components/schemas/{SchemaName}')
				refPath := v
				if len(refPath) > 0 && refPath[0] == '#' {
					refName := filepath.Base(refPath) // Get the schema name from the path
					referencedSchema, ok := schemas[refName]
					if !ok {
						return nil, fmt.Errorf("referenced schema %s not found", refName)
					}

					// Replace $ref with the actual schema definition
					for k, v := range referencedSchema.(map[string]interface{}) {
						currentSchema[k] = v
					}

					// Remove $ref key after resolving
					delete(currentSchema, "$ref")
				} else {
					// Handle external refs (could be expanded to handle external files, if needed)
					return nil, fmt.Errorf("external $ref not supported: %s", refPath)
				}
			}
		}

		if key == "nullable" && value.(bool) {
			// Check the current type of the schema
			if currentType, ok := currentSchema["type"]; ok {
				switch t := currentType.(type) {
				case string:
					// Set type to a slice with the original type and "null"
					currentSchema["type"] = []interface{}{t, "null"}
				case []interface{}:
					// Check if "null" is already in the array
					hasNull := false
					for _, v := range t {
						if v == "null" {
							hasNull = true
							break
						}
					}
					// Only append "null" if it's not already present
					if !hasNull {
						currentSchema["type"] = append(t, "null")
					}
				default:
					// In case type is unknown, default to ["object", "null"]
					currentSchema["type"] = []interface{}{"object", "null"}
				}
			}
		}

		// If it's an object, add "additionalProperties: false"
		if key == "type" && value == "object" {
			if _, exists := currentSchema["additionalProperties"]; !exists {
				currentSchema["additionalProperties"] = false
			}
		}
	}

	return currentSchema, nil
}

// convertToJSON converts a generic map to a JSON string
func convertToJSON(data map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Validate validates the API response against a specified schema from the OpenAPI file
func Validate(openApiFilePath string, schemaName string, apiResponse io.ReadCloser) (bool, error) {
	defer apiResponse.Close()

	// Step 1: Read and parse the OpenAPI YAML file
	openAPISpec, err := readYAMLFile(openApiFilePath)
	if err != nil {
		return false, fmt.Errorf("error reading OpenAPI spec: %w", err)
	}

	// Step 2: Resolve $ref fields within the schema
	basePath := filepath.Dir(openApiFilePath) // Base path for resolving local files
	resolvedSchema, err := resolveRefs(openAPISpec, basePath)
	if err != nil {
		return false, fmt.Errorf("error resolving $ref references: %w", err)
	}

	// Step 3: Extract the fully resolved schema by name (e.g., "Source" or "SourcesList")
	schema, ok := resolvedSchema[schemaName]
	if !ok {
		return false, fmt.Errorf("schema %s not found", schemaName)
	}
	// Step 4: Convert the extracted schema to JSON
	schemaJSON, err := convertToJSON(schema.(map[string]interface{}))
	if err != nil {
		return false, fmt.Errorf("error converting schema to JSON: %w", err)
	}

	// Step 5: Read the entire API response from the io.ReadCloser
	responseBytes, err := io.ReadAll(apiResponse)
	if err != nil {
		return false, fmt.Errorf("error reading API response: %w", err)
	}

	// Step 6: Validate the API response against the schema using gojsonschema
	schemaLoader := gojsonschema.NewStringLoader(schemaJSON)
	responseLoader := gojsonschema.NewBytesLoader(responseBytes)

	// Perform validation
	result, err := gojsonschema.Validate(schemaLoader, responseLoader)
	if err != nil {
		return false, fmt.Errorf("error during validation: %w", err)
	}

	// Step 7: Check validation result and return
	if !result.Valid() {
		// Collect validation errors in a string
		var errorMessages string
		for _, desc := range result.Errors() {
			errorMessages += fmt.Sprintf("- %s\n", desc)
		}
		return false, fmt.Errorf("validation failed: \n%s", errorMessages)
	}

	return true, nil
}
