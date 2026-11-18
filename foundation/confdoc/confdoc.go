// Package confdoc describes the fields of a configuration file once, and
// renders that description as help text and as a JSON Schema.
//
// The description is data rather than struct tags or doc comments: a comment
// is invisible to the running program, and a tag full of prose is a string
// nobody formats. One table, read by the help, the schema and the starter
// configuration alike, is what keeps the three from disagreeing.
package confdoc

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Kind is the JSON shape of a field.
type Kind string

const (
	String  Kind = "string"
	Boolean Kind = "boolean"
	Integer Kind = "integer"
	// Strings is an array of strings.
	Strings Kind = "array of strings"
	// StringMap is an object whose values are strings.
	StringMap Kind = "object of strings"
	// AnyMap is an object whose values may be anything.
	AnyMap Kind = "object"
	// Object is a nested object whose fields are listed separately, with this
	// field's path as their prefix.
	Object Kind = "object with fields"
	// Objects is an array of objects whose fields are listed separately.
	Objects Kind = "array of objects"
)

// Field describes one field.
type Field struct {
	// Path is the dotted path, e.g. "icon.path". Fields of the items of an
	// array of objects use the array's path as prefix: "binaries.arch".
	Path        string   `json:"path"`
	Kind        Kind     `json:"kind"`
	Description string   `json:"description"`
	Default     string   `json:"default,omitempty"`
	Example     string   `json:"example,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Required    bool     `json:"required,omitempty"`
	// Flag names the command line option that overrides this field.
	Flag string `json:"flag,omitempty"`
}

// Schema renders a JSON Schema (draft 2020-12) of the fields.
//
// Every object refuses properties it does not list, because the reader of the
// file refuses them too; a schema that allowed them would let an editor accept
// a typo the tool then rejects.
func Schema(id, title, description string, fields []Field) []byte {
	root := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"$id":                  id,
		"title":                title,
		"description":          description,
		"type":                 "object",
		"additionalProperties": false,
	}
	props := map[string]any{
		"$schema": map[string]any{"type": "string", "description": "The schema this file follows, for editors. Ignored by the tool."},
	}
	required := childrenInto(props, "", fields)
	root["properties"] = props
	if len(required) > 0 {
		root["required"] = required
	}
	out, _ := json.MarshalIndent(root, "", "  ")
	return append(out, '\n')
}

// childrenInto fills props with the direct children of prefix and returns the
// names of the required ones.
func childrenInto(props map[string]any, prefix string, fields []Field) []string {
	var required []string
	for _, f := range fields {
		parent, name := split(f.Path)
		if parent != prefix {
			continue
		}
		props[name] = node(f, fields)
		if f.Required {
			required = append(required, name)
		}
	}
	slices.Sort(required)
	return required
}

func split(path string) (string, string) {
	i := strings.LastIndex(path, ".")
	if i < 0 {
		return "", path
	}
	return path[:i], path[i+1:]
}

func node(f Field, fields []Field) map[string]any {
	n := map[string]any{"description": f.Description}
	if f.Example != "" {
		n["examples"] = []any{example(f)}
	}
	switch f.Kind {
	case String:
		n["type"] = "string"
		if len(f.Enum) > 0 {
			n["enum"] = f.Enum
		}
	case Boolean:
		n["type"] = "boolean"
	case Integer:
		n["type"] = "integer"
	case Strings:
		n["type"] = "array"
		n["items"] = map[string]any{"type": "string"}
	case StringMap:
		n["type"] = "object"
		n["additionalProperties"] = map[string]any{"type": "string"}
	case AnyMap:
		n["type"] = "object"
	case Object, Objects:
		obj := map[string]any{"type": "object", "additionalProperties": false}
		props := map[string]any{}
		if req := childrenInto(props, f.Path, fields); len(req) > 0 {
			obj["required"] = req
		}
		obj["properties"] = props
		if f.Kind == Object {
			for k, v := range obj {
				n[k] = v
			}
		} else {
			n["type"] = "array"
			n["items"] = obj
		}
	}
	return n
}

// example returns the example as the JSON value it stands for, falling back to
// the text when it is not valid JSON on its own.
func example(f Field) any {
	var v any
	if err := json.Unmarshal([]byte(f.Example), &v); err == nil {
		return v
	}
	return f.Example
}

// Text renders the fields as help text.
func Text(fields []Field, wrap func(text string, indent int) string) string {
	var b strings.Builder
	for i, f := range fields {
		if i > 0 {
			b.WriteString("\n")
		}
		head := f.Path + "  (" + string(f.Kind)
		if f.Required {
			head += ", required"
		}
		head += ")"
		b.WriteString(head + "\n")
		var extra []string
		if len(f.Enum) > 0 {
			extra = append(extra, "One of: "+strings.Join(f.Enum, ", ")+".")
		}
		if f.Default != "" {
			extra = append(extra, "Default: "+f.Default+".")
		}
		if f.Flag != "" {
			extra = append(extra, "Option: --"+f.Flag+".")
		}
		if f.Example != "" {
			extra = append(extra, "Example: "+f.Example)
		}
		b.WriteString(wrap(strings.TrimSpace(f.Description+" "+strings.Join(extra, " ")), 4))
	}
	return b.String()
}

// Names returns the field paths, for suggestions.
func Names(fields []Field) []string {
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, f.Path)
	}
	return out
}

// Check reports fields whose parent is not declared as an object, so that a
// misspelt prefix in the table is found by a test rather than by a user.
func Check(fields []Field) error {
	kinds := map[string]Kind{}
	for _, f := range fields {
		kinds[f.Path] = f.Kind
	}
	for _, f := range fields {
		parent, _ := split(f.Path)
		if parent == "" {
			continue
		}
		if k := kinds[parent]; k != Object && k != Objects {
			return fmt.Errorf("field %s has parent %s, which is not an object", f.Path, parent)
		}
	}
	return nil
}
