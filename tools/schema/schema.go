package schema

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/codebahn/codebahn-cli/tools"
)

// For generates a JSON Schema object from a ToolDef's Args struct.
// The schema is a flat {"type":"object","properties":{...},"required":[...]}
// derived from the struct's field tags.
func For(td tools.ToolDef) json.RawMessage {
	rt := reflect.TypeOf(td.Args)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	properties := map[string]map[string]any{}
	var required []string

	for i := range rt.NumField() {
		f := rt.Field(i)
		name := f.Tag.Get("json")
		if name == "" || name == "-" {
			continue
		}

		prop := map[string]any{
			"type":        goTypeToJSONType(f.Type),
			"description": f.Tag.Get("desc"),
		}

		if def := f.Tag.Get("default"); def != "" {
			prop["default"] = coerceDefault(def, f.Type)
		}

		properties[name] = prop

		if f.Tag.Get("required") == "true" {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"type": "object",
	}
	if len(properties) > 0 {
		schema["properties"] = properties
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	raw, _ := json.Marshal(schema)
	return raw
}

func goTypeToJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	default:
		return "string"
	}
}

func coerceDefault(val string, t reflect.Type) any {
	switch t.Kind() {
	case reflect.Bool:
		return val == "true"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			return n
		}
		return val
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return val
	default:
		return val
	}
}
