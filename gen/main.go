//go:build ignore

// Command gen regenerates the JSON Schemas embedded by the schema
// subpackage from their Go-first source of truth: scheme.Profile's field
// documentation becomes each property's "description".
//
// Run from the module root:
//
//	go run gen/main.go
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/plexusone/resourceref/scheme"
)

func main() {
	generateProfileSchema()
	generateResourceRefSchema()
}

func generateProfileSchema() {
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/plexusone/resourceref", "./scheme"); err != nil {
		log.Fatalf("gen: reading scheme package comments: %v", err)
	}
	s := r.Reflect(&scheme.Profile{})
	s.ID = "https://raw.githubusercontent.com/plexusone/resourceref/main/schema/profile.schema.json"
	s.Title = "resourceref Scheme Profile"
	s.Description = "Data-driven definition of an owned URI scheme (resourceref/scheme.Profile). " +
		"A namespace owner's own scheme instance — for example PlexusOne's \"prn\" scheme — " +
		"is a document conforming to this schema, defined in that owner's standards repository."
	writeSchema("schema/profile.schema.json", s)
}

// generateResourceRefSchema writes the fragment other JSON Schemas embed
// for a field of type resourceref.ResourceRef. It is not reflected from
// the Go type — ResourceRef marshals as a plain string via
// encoding.TextMarshaler, and reflecting its unexported internal field
// would be meaningless — so this is a small, hand-composed literal using
// only encoding/json, keeping the generator itself free of any dependency
// beyond invopop/jsonschema.
func generateResourceRefSchema() {
	s := map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"$id":         "https://raw.githubusercontent.com/plexusone/resourceref/main/schema/resourceref.schema.json",
		"title":       "ResourceRef",
		"description": "A resourceref.ResourceRef, encoded as its canonical URI string (RFC 3986). Embed this fragment via $ref for any JSON property typed as a resource reference.",
		"type":        "string",
		"format":      "uri",
		"minLength":   1,
	}
	writeSchemaMap("schema/resourceref.schema.json", s)
}

func writeSchema(path string, s *jsonschema.Schema) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatalf("gen: marshaling %s: %v", path, err)
	}
	writeFile(path, data)
}

func writeSchemaMap(path string, s map[string]any) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatalf("gen: marshaling %s: %v", path, err)
	}
	writeFile(path, data)
}

func writeFile(path string, data []byte) {
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		log.Fatalf("gen: writing %s: %v", path, err)
	}
	log.Printf("gen: wrote %s", path)
}
