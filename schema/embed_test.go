package schema

import (
	"encoding/json"
	"testing"
)

func TestEmbeddedSchemasAreValidJSON(t *testing.T) {
	for name, data := range map[string][]byte{
		"ProfileSchema":     ProfileSchema,
		"ResourceRefSchema": ResourceRefSchema,
	} {
		t.Run(name, func(t *testing.T) {
			if len(data) == 0 {
				t.Fatalf("%s is empty", name)
			}
			var v map[string]any
			if err := json.Unmarshal(data, &v); err != nil {
				t.Fatalf("%s is not valid JSON: %v", name, err)
			}
			if v["$schema"] == "" || v["$schema"] == nil {
				t.Fatalf("%s has no $schema", name)
			}
		})
	}
}

func TestProfileSchemaReferencesProfile(t *testing.T) {
	var v map[string]any
	if err := json.Unmarshal(ProfileSchema, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v["$ref"] != "#/$defs/Profile" {
		t.Fatalf("ProfileSchema $ref = %v, want #/$defs/Profile", v["$ref"])
	}
	defs, ok := v["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("ProfileSchema has no $defs object")
	}
	for _, want := range []string{"Profile", "Authority", "Shape", "SegmentSpec", "ResourceTypeRegistry", "URISubset", "VersionGrammar"} {
		if _, ok := defs[want]; !ok {
			t.Errorf("ProfileSchema $defs missing %q", want)
		}
	}
}

func TestResourceRefSchemaShape(t *testing.T) {
	var v map[string]any
	if err := json.Unmarshal(ResourceRefSchema, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v["type"] != "string" {
		t.Fatalf(`ResourceRefSchema["type"] = %v, want "string"`, v["type"])
	}
	if v["format"] != "uri" {
		t.Fatalf(`ResourceRefSchema["format"] = %v, want "uri"`, v["format"])
	}
}
