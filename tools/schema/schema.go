package schema

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/codebahn/codebahn-cli/tools"
)

// ForArgs generates a JSON Schema from an arbitrary args struct.
func ForArgs(args any) json.RawMessage {
	return forStruct(args)
}

// For generates a JSON Schema object from a ToolDef's Args struct.
// The schema is a flat {"type":"object","properties":{...},"required":[...]}
// derived from the struct's field tags.
func For(td tools.ToolDef) json.RawMessage {
	return forStruct(td.Args)
}

func forStruct(args any) json.RawMessage {
	rt := reflect.TypeOf(args)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	properties := map[string]map[string]any{}
	var required []string

	for i := range rt.NumField() {
		f := rt.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
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

		if enum := f.Tag.Get("enum"); enum != "" {
			vals := strings.Split(enum, ",")
			enumVals := make([]any, len(vals))
			for i, v := range vals {
				enumVals[i] = v
			}
			prop["enum"] = enumVals
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
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
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
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
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
