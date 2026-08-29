//go:build tools

// This file has no build purpose beyond keeping `go mod tidy` from
// stripping the schema generator's dependency out of go.mod, or marking it
// indirect: `go mod tidy` scans build-tag-gated files (including this
// "tools"-tagged one) when computing the module's full possible import
// graph, so this blank import is enough to keep invopop/jsonschema listed
// as a direct requirement even though the only other importer, main.go, is
// tagged `//go:build ignore` and is never itself part of a normal build.
// Regenerate the schema with:
//
//	go run gen/main.go
package main

import (
	_ "github.com/invopop/jsonschema"
)
