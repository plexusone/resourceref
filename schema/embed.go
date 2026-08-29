// Package schema embeds the JSON Schemas generated from this module's
// Go-first source of truth (see the gen/ directory), for external tooling
// (schemakit, documentation generators, non-Go consumers) and for other
// JSON Schema documents that want to $ref a resource-reference field or a
// scheme profile document.
//
// This package is not used by resourceref's own runtime validation
// (Load's structural checks are plain Go, per the module's stdlib-only
// constraint) — it exists so other schemas can point at these files rather
// than re-deriving the same shape.
package schema

import _ "embed"

// ProfileSchema is the generated JSON Schema for a [scheme.Profile]
// document — what a namespace owner's own scheme instance (for example
// PlexusOne's "prn" scheme, defined in plexusone-standards) must conform
// to.
//
//go:embed profile.schema.json
var ProfileSchema []byte

// ResourceRefSchema is the JSON Schema fragment for a JSON property typed
// as a [resourceref.ResourceRef]: a non-empty string in canonical URI
// form. Embed it via $ref rather than restating "type: string, format:
// uri" in every schema that has a reference field.
//
//go:embed resourceref.schema.json
var ResourceRefSchema []byte
