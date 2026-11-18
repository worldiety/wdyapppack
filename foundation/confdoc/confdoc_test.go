package confdoc

import (
	"encoding/json"
	"strings"
	"testing"
)

func fields() []Field {
	return []Field{
		{Path: "name", Kind: String, Required: true, Description: "the name", Example: `"x"`},
		{Path: "items", Kind: Objects, Description: "items"},
		{Path: "items.path", Kind: String, Required: true, Description: "a path"},
		{Path: "icon", Kind: Object, Description: "icon"},
		{Path: "icon.shape", Kind: String, Enum: []string{"a", "b"}, Description: "shape"},
		{Path: "env", Kind: StringMap, Description: "env"},
	}
}

func TestSchemaNestsAndRefusesUnknownFields(t *testing.T) {
	var s map[string]any
	if err := json.Unmarshal(Schema("id", "t", "d", fields()), &s); err != nil {
		t.Fatal(err)
	}
	if s["additionalProperties"] != false || s["required"].([]any)[0] != "name" {
		t.Fatalf("%v", s)
	}
	props := s["properties"].(map[string]any)
	items := props["items"].(map[string]any)["items"].(map[string]any)
	if items["additionalProperties"] != false || items["required"].([]any)[0] != "path" {
		t.Fatalf("%v", items)
	}
	shape := props["icon"].(map[string]any)["properties"].(map[string]any)["shape"].(map[string]any)
	if len(shape["enum"].([]any)) != 2 {
		t.Fatalf("%v", shape)
	}
}

func TestCheckFindsOrphans(t *testing.T) {
	if err := Check(fields()); err != nil {
		t.Fatal(err)
	}
	if err := Check(append(fields(), Field{Path: "nope.x", Kind: String})); err == nil {
		t.Fatal("an orphan was accepted")
	}
}

func TestText(t *testing.T) {
	out := Text(fields(), func(s string, indent int) string { return strings.Repeat(" ", indent) + s + "\n" })
	if !strings.Contains(out, "icon.shape  (string)\n    shape One of: a, b.") {
		t.Fatalf("%s", out)
	}
}
