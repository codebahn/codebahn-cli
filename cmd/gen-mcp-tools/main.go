// Command gen-mcp-tools writes the consolidated MCP tool registry as JSON to stdout.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/codebahn/codebahn-cli/tools"
)

type toolJSON struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Actions     []string    `json:"actions"`
	Params      []paramJSON `json:"params"`
}

type paramJSON struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Desc     string   `json:"desc"`
	Default  string   `json:"default,omitempty"`
	Enum     []string `json:"enum,omitempty"`
}

func main() {
	var out []toolJSON

	for _, mt := range tools.MCPAll {
		t := toolJSON{
			Name:        mt.Name,
			Description: mt.Description,
			Actions:     deduplicateActions(mt.Actions),
			Params:      extractParams(mt.Args),
		}
		out = append(out, t)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func deduplicateActions(actions []tools.MCPAction) []string {
	seen := map[string]bool{}
	out := []string{} // non-nil so JSON encodes as [] not null
	for _, a := range actions {
		if a.Action == "" || seen[a.Action] {
			continue
		}
		seen[a.Action] = true
		out = append(out, a.Action)
	}
	return out
}

func extractParams(args any) []paramJSON {
	rt := reflect.TypeOf(args)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.NumField() == 0 {
		return []paramJSON{}
	}

	// Collect required field names for sorting required-first.
	reqSet := map[string]bool{}
	for i := range rt.NumField() {
		f := rt.Field(i)
		if f.Tag.Get("required") == "true" {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			reqSet[name] = true
		}
	}

	var params []paramJSON
	for i := range rt.NumField() {
		f := rt.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}

		p := paramJSON{
			Name:     name,
			Type:     goTypeToJSONType(f.Type),
			Required: f.Tag.Get("required") == "true",
			Desc:     f.Tag.Get("desc"),
			Default:  f.Tag.Get("default"),
		}

		if enum := f.Tag.Get("enum"); enum != "" {
			p.Enum = strings.Split(enum, ",")
		}

		params = append(params, p)
	}

	// Stable sort: required params first, then original order within each group.
	sort.SliceStable(params, func(i, j int) bool {
		ri, rj := params[i].Required, params[j].Required
		if ri != rj {
			return ri
		}
		return false
	})

	return params
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
