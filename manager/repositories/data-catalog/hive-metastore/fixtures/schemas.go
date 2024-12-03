package fixtures

import (
	_ "embed"
	"encoding/json"

	"github.com/beltran/gohive/hive_metastore"
)

const (
	simpleJsonSchema = `{
    "type": "object",
    "properties": {
      "column_bol": {
        "type": "boolean"
      },
      "column_str": {
        "type": "string"
      },
      "column_str.period": {
        "type": "string"
      },
      "column_date": {
        "type": "string",
        "format": "date"
      },
      "column_datetime": {
        "type": "string",
        "format": "date-time",
        "airbyte_type": "timestamp_without_timezone"
      },
      "column_time": {
        "type": "string",
        "format": "time",
        "airbyte_type": "time_without_timezone"
      },
      "column_int": {
        "type": "integer"
      },
      "column_num": {
        "type": "number"
      }
    }
  }`
	nestedJsonSchema = `{
    "type": "object",
    "properties": {
      "column_str": {
        "type": "string"
      },
      "column_str_null": {
        "type": [
          "null",
          "string"
        ]
      },
      "column_obj": {
        "type": "object",
        "properties": {
          "column_inner_obj": {
            "type": "object",
            "properties": {
              "column_str": {
                "type": "string"
              }
            }
          }
        }
      },
      "column_str_arr": {
        "type": "array",
        "items": {
          "type": "string"
        }
      },
      "column_obj_arr": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "column_1": {
              "type": "string"
            },
            "column_2": {
              "type": "integer"
            }
          }
        }
      },
      "column_bug_arr": {
        "type": "array"
      },
      "column_bug_obj": {
        "type": "object"
      },
      "column_composite_key": {
        "anyOf": [
          {
            "type": "object",
            "properties": {
              "column_str_1": {
                "type": "string"
              }
            }
          },
          {
            "type": "object",
            "properties": {
              "column_str_2": {
                "type": "string"
              }
            }
          }
        ]
      }
    }
  }`
)

func GetSimpleJsonSchemaFixtures() (map[string]interface{}, []*hive_metastore.FieldSchema) {
	jsonSchema := map[string]interface{}{}
	json.Unmarshal([]byte(simpleJsonSchema), &jsonSchema)

	cols := []*hive_metastore.FieldSchema{
		&hive_metastore.FieldSchema{
			Name: "column_bol",
			Type: "boolean",
		},
		&hive_metastore.FieldSchema{
			Name: "column_date",
			Type: "date",
		},
		&hive_metastore.FieldSchema{
			Name: "column_datetime",
			Type: "timestamp",
		},
		&hive_metastore.FieldSchema{
			Name: "column_int",
			Type: "bigint",
		},
		&hive_metastore.FieldSchema{
			Name: "column_num",
			Type: "double",
		},
		&hive_metastore.FieldSchema{
			Name: "column_str",
			Type: "string",
		},
		&hive_metastore.FieldSchema{
			Name: "column_str_period",
			Type: "string",
		},
		&hive_metastore.FieldSchema{
			Name: "column_time",
			Type: "bigint",
		},
	}
	return jsonSchema, cols
}

func GetNestedJsonSchemaFixtures() (map[string]interface{}, []*hive_metastore.FieldSchema) {
	jsonSchema := map[string]interface{}{}
	json.Unmarshal([]byte(nestedJsonSchema), &jsonSchema)

	cols := []*hive_metastore.FieldSchema{
		&hive_metastore.FieldSchema{
			Name: "column_bug_arr",
			Type: "error",
		},
		&hive_metastore.FieldSchema{
			Name: "column_bug_obj",
			Type: "error",
		},
		&hive_metastore.FieldSchema{
			Name: "column_composite_key",
			Type: "struct<column_str_1:string,column_str_2:string>",
		},
		&hive_metastore.FieldSchema{
			Name: "column_obj",
			Type: "struct<column_inner_obj:struct<column_str:string>>",
		},
		&hive_metastore.FieldSchema{
			Name: "column_obj_arr",
			Type: "array<struct<column_1:string,column_2:bigint>>",
		},
		&hive_metastore.FieldSchema{
			Name: "column_str",
			Type: "string",
		},
		&hive_metastore.FieldSchema{
			Name: "column_str_arr",
			Type: "array<string>",
		},
		&hive_metastore.FieldSchema{
			Name: "column_str_null",
			Type: "string",
		},
	}
	return jsonSchema, cols
}

//go:embed hubspot_deals.json
var hubspotDealsJsonSchema []byte

func GetHubspotDealsJsonSchemaFixtures() (map[string]interface{}, int) {
	jsonSchema := map[string]interface{}{}
	json.Unmarshal(hubspotDealsJsonSchema, &jsonSchema)

	return jsonSchema, 278
}

//go:embed shopify_orders.json
var shopifyOrdersJsonSchema []byte

func GetShopifyOrdersJsonSchemaFixtures() (map[string]interface{}, int) {
	jsonSchema := map[string]interface{}{}
	json.Unmarshal(shopifyOrdersJsonSchema, &jsonSchema)

	return jsonSchema, 94
}
