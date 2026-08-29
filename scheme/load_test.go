package scheme

import (
	"errors"
	"testing"
)

// validMinimalProfile is a small, well-formed profile used as a base for
// mutation-based negative tests: each invalid case starts from this JSON
// and breaks exactly one thing.
const validMinimalProfile = `{
	"scheme": "prn",
	"uriSubset": {
		"allowQuery": false,
		"allowFragment": false,
		"allowUserinfo": false,
		"allowPort": false,
		"segmentPattern": "^[a-z][a-z0-9_-]*$"
	},
	"authorities": [
		{"name": "acme", "description": "Acme Corp"},
		{"name": "opensource"}
	],
	"shapes": [
		{
			"name": "tenant-scoped",
			"segments": [
				{"role": "app"},
				{"role": "org"},
				{"role": "resourceType"},
				{"role": "id"}
			]
		},
		{
			"name": "ecosystem-artifact",
			"segments": [
				{"role": "resourceType"},
				{"role": "id", "rest": true}
			]
		}
	],
	"resourceTypes": [
		{"authority": "acme", "app": "billing", "types": ["invoice", "shipment"]}
	],
	"version": {
		"separator": "@",
		"required": true,
		"versionPattern": "^[0-9]+$",
		"aliasPattern": "^[a-z][a-z0-9-]*$"
	}
}`

func TestLoadValidProfile(t *testing.T) {
	p, err := Load([]byte(validMinimalProfile))
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if p.Scheme != "prn" {
		t.Fatalf("Scheme = %q, want %q", p.Scheme, "prn")
	}
	if len(p.Authorities) != 2 {
		t.Fatalf("len(Authorities) = %d, want 2", len(p.Authorities))
	}
	if len(p.Shapes) != 2 {
		t.Fatalf("len(Shapes) = %d, want 2", len(p.Shapes))
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	_, err := Load([]byte(`{not json`))
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("Load(malformed JSON): err = %v, want wrapping ErrInvalidProfile", err)
	}
}

func TestLoadStructuralErrors(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{
			name: "empty scheme",
			json: `{"scheme":"","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}]}`,
		},
		{
			name: "scheme starting with a digit",
			json: `{"scheme":"1bad","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}]}`,
		},
		{
			name: "no authorities",
			json: `{"scheme":"prn","authorities":[],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}]}`,
		},
		{
			name: "authority with empty name",
			json: `{"scheme":"prn","authorities":[{"name":""}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}]}`,
		},
		{
			name: "duplicate authority name",
			json: `{"scheme":"prn","authorities":[{"name":"a"},{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}]}`,
		},
		{
			name: "no shapes",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[]}`,
		},
		{
			name: "shape with no segments",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[]}]}`,
		},
		{
			name: "shape with unknown role",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"bogus"}]}]}`,
		},
		{
			name: "shape missing a resourceType segment",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"id"}]}]}`,
		},
		{
			name: "rest on a non-id role",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType","rest":true},{"role":"id"}]}]}`,
		},
		{
			name: "rest segment not last",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"id","rest":true},{"role":"resourceType"}]}]}`,
		},
		{
			name: "duplicate role within one shape",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"},{"role":"resourceType"}]}]}`,
		},
		{
			name: "two rest segments",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"},{"role":"id","rest":true}]},{"name":"s2","segments":[{"role":"resourceType"},{"role":"id","rest":true},{"role":"id","rest":true}]}]}`,
		},
		{
			name: "resourceType registry references unregistered authority",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}],"resourceTypes":[{"authority":"nope","types":["x"]}]}`,
		},
		{
			name: "resourceType registry with no types",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}],"resourceTypes":[{"authority":"a","types":[]}]}`,
		},
		{
			name: "invalid segmentPattern regex",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}],"uriSubset":{"segmentPattern":"(unclosed"}}`,
		},
		{
			name: "version required with no pattern",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}],"version":{"required":true}}`,
		},
		{
			name: "invalid versionPattern regex",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}],"version":{"separator":"@","versionPattern":"(unclosed"}}`,
		},
		{
			name: "invalid aliasPattern regex",
			json: `{"scheme":"prn","authorities":[{"name":"a"}],"shapes":[{"name":"s","segments":[{"role":"resourceType"}]}],"version":{"separator":"@","aliasPattern":"(unclosed"}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load([]byte(tc.json))
			if !errors.Is(err, ErrInvalidProfile) {
				t.Fatalf("Load: err = %v, want wrapping ErrInvalidProfile", err)
			}
		})
	}
}
