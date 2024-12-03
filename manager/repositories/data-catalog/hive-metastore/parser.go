package hivemetastore

import (
	"fmt"
	"regexp"
	"strings"

	"dario.cat/mergo"
	"github.com/beltran/gohive/hive_metastore"
)

const (
	typeKey        = "type"
	airbyteTypeKey = "airbyte_type"
	anyOfKey       = "anyOf"
	allOfKey       = "allOf"
	oneOfKey       = "oneOf"
	formatKey      = "format"
	propertiesKey  = "properties"
	itemsKey       = "items"
	hiveStringType = "string"
	errorType      = "error"
)

var (
	jsonSchemaTypeToHiveType = map[string]string{
		"integer":                    "bigint",
		"number":                     "double",
		"string":                     hiveStringType,
		"date":                       "date",
		"date-time":                  "timestamp",
		"time":                       "bigint",
		"time_with_timezone":         "bigint",
		"time_without_timezone":      "bigint",
		"timestamp_with_timezone":    "timestamp",
		"timestamp_without_timezone": "timestamp",
		"binary":                     "binary",
		"boolean":                    "boolean",
	}
	compositionKeys = map[string]bool{
		anyOfKey: true,
		allOfKey: true,
		oneOfKey: true,
	}
)

// We need to parse the JSON Schema to the Hive Metastore format
// while keeping in mind that Airbyte parses the JSON Schema to Avro
// and then it generates the Parquet record based on the Avro format
// for reference and future support read
// https://docs.airbyte.com/understanding-airbyte/json-avro-conversion
func getTypeRecursive(schema map[string]interface{}) string {
	schemaTypeInterface := schema[typeKey]
	schemaType := ""
	switch schemaTypeInterface.(type) {
	case []interface{}:
		schemaType = schemaTypeInterface.([]interface{})[1].(string)
	case string:
		schemaType = schema[typeKey].(string)
	default:
		for key := range compositionKeys {
			if _, exists := schema[key]; exists {
				schemaType = key
			}
		}
	}

	switch schemaType {
	case "object":
		if properties, exists := schema[propertiesKey]; exists {
			nestedProperties := properties.(map[string]interface{})
			nestedTypes := []string{}
			for nestedColName, nestedValue := range nestedProperties {
				recursiveType := getTypeRecursive(nestedValue.(map[string]interface{}))
				nestedTypes = append(nestedTypes, fmt.Sprintf("%s:%s", nestedColName, recursiveType))
			}
			return fmt.Sprintf("struct<%s>", strings.Join(nestedTypes, ","))
		} else {
			// this should not happen so we will return error
			return errorType
		}
	case "array":
		if items, exists := schema[itemsKey]; exists {
			nestedType := getTypeRecursive(items.(map[string]interface{}))
			return fmt.Sprintf("array<%s>", nestedType)
		} else {
			// this should not happen so we will return error
			return errorType
		}
	default:
		if _, isCompositeKey := compositionKeys[schemaType]; isCompositeKey {
			if values, exists := schema[schemaType]; exists {
				mergedTypes := map[string]interface{}{}
				for _, value := range values.([]interface{}) {
					if err := mergo.Merge(&mergedTypes, value); err != nil {
						return "error"
					}
				}
				fmt.Println(mergedTypes)
				return getTypeRecursive(mergedTypes)
			}
		} else {
			if airbyteType, exists := schema[airbyteTypeKey]; exists {
				schemaType = airbyteType.(string)
			} else if format, exists := schema[formatKey]; exists {
				schemaType = format.(string)
			}
		}
		if jsonType, exists := jsonSchemaTypeToHiveType[schemaType]; exists {
			return jsonType
		}
		// this should not happen so we will return error
		return errorType
	}
}

const (
	cleanHiveColumnNameRegex = `[\.]`
	columnNameCharReplace    = "_"
)

func cleanColumnName(name string) string {
	cleanHiveColumnNameRegexCompiled := regexp.MustCompile(cleanHiveColumnNameRegex)
	return cleanHiveColumnNameRegexCompiled.ReplaceAllString(name, columnNameCharReplace)
}

// TODO: load using jsonschema lib later for better type assert
// TODO: add description from json schema
func GetColumnsFromJsonSchema(jsonSchema map[string]interface{}) []*hive_metastore.FieldSchema {
	schema := jsonSchema[propertiesKey].(map[string]interface{})
	cols := make([]*hive_metastore.FieldSchema, len(schema))

	i := 0
	for colName, val := range schema {
		hiveColName := cleanColumnName(colName)
		col := hive_metastore.FieldSchema{Name: hiveColName}
		col.Type = getTypeRecursive(val.(map[string]interface{}))
		cols[i] = &col
		i += 1
	}

	return cols
}
