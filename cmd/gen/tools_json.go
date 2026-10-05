package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"

	"github.com/codebahn/codebahn-cli/tools"
	"github.com/codebahn/codebahn-cli/tools/schema"
)

// toolsJSONParam is one parameter of a tools.json entry.
type toolsJSONParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Desc     string `json:"desc"`
}

// toolsJSONEntry is one entry in the generated tools.json tool reference.
type toolsJSONEntry struct {
	Name        string           `json:"name"`
	Group       string           `json:"group"`
	Description string           `json:"description"`
	Method      string           `json:"method"`
	Destructive bool             `json:"destructive"`
	ReadOnly    bool             `json:"read_only"`
	Params      []toolsJSONParam `json:"params"`
}

// buildToolsJSON builds the tools.json entries from tools.All, preserving
// registry order.
func buildToolsJSON() []toolsJSONEntry {
	entries := make([]toolsJSONEntry, len(tools.All))
	for i, td := range tools.All {
		entries[i] = toolsJSONEntry{
			Name:        td.Name,
			Group:       td.Group,
			Description: td.Description,
			Method:      td.Method,
			Destructive: td.Destructive,
			ReadOnly:    td.IsReadOnly(),
			Params:      buildToolsJSONParams(td),
		}
	}
	return entries
}

// buildToolsJSONParams extracts params for one tool in struct declaration
// order. Types come from schema.For (which shares the type vocabulary), but
// schema.For emits properties alphabetically, so field order and the
// required/skip rules are taken from the struct tags directly.
func buildToolsJSONParams(td tools.ToolDef) []toolsJSONParam {
	var parsed struct {
		Properties map[string]struct {
			Type string `json:"type"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(schema.For(td), &parsed); err != nil {
		panic(err)
	}
	types := parsed.Properties

	rt := reflect.TypeOf(td.Args)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	params := []toolsJSONParam{}
	for i := range rt.NumField() {
		f := rt.Field(i)
		name := f.Tag.Get("json")
		if name == "" || name == "-" {
			continue
		}

		params = append(params, toolsJSONParam{
			Name:     name,
			Type:     types[name].Type,
			Required: f.Tag.Get("required") == "true",
			Desc:     f.Tag.Get("desc"),
		})
	}
	return params
}

// marshalToolsJSON renders entries as the committed tools.json: 2-space
// indent, no HTML escaping (descriptions can contain `&`, `<`, `>`
// unescaped), trailing newline.
func marshalToolsJSON(entries []toolsJSONEntry) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeToolsJSON builds and writes the tools.json tool reference to path.
func writeToolsJSON(path string) error {
	data, err := marshalToolsJSON(buildToolsJSON())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
