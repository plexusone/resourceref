# Embedding resourceref in an IR

This guide is for anyone authoring a JSON IR or Go type — a build
artifact, a threat-model document, a telemetry event, a service's API
schema — that needs to reference another resource: a repo, a service, an
agent, a workflow run, a ticket. It covers three things: how to add a
reference field, how to represent one in a JSON Schema, and how to avoid
the identity mistake this module exists to prevent.

## The one rule

> A resource reference is a URI. Prefer the resource owner's canonical URI
> when one exists — an AWS ARN, a SPIFFE ID, a GitHub URL — and mint an
> owned URI only when one does not.

Don't invent a `resource_id` string convention for your IR. If the thing
you're referencing already has a canonical URI (a GitHub repo, an AWS
resource, a SPIFFE workload identity), use that URI directly. Only reach
for an owned scheme — one you define yourself with the `scheme` package,
this module ships no scheme of its own — when the resource has no other
canonical identity.

## Adding a reference field to a Go type

```go
import "github.com/plexusone/resourceref"

type BuildArtifact struct {
	ID          string                    `json:"id"`
	SourceRepo  resourceref.ResourceRef   `json:"sourceRepo"`
	DeployedTo  *resourceref.ResourceRef  `json:"deployedTo,omitempty"`
}
```

- Construct references with `resourceref.Parse`, never by formatting a
  string yourself — canonicalization is what makes two references
  comparable.
- `ResourceRef` implements `encoding.TextMarshaler`/`TextUnmarshaler`, so
  it encodes as a plain JSON string with no extra code:

  ```json
  {"id": "art_1", "sourceRepo": "https://github.com/acme/billing-service"}
  ```

- **A required reference is `ResourceRef`; an optional one is
  `*ResourceRef` with `omitempty`.** The zero `ResourceRef` is invalid by
  construction (you can only get a valid one from `Parse`) and
  deliberately errors on marshal rather than silently emitting `""` — so
  don't use the zero value to mean "no reference." Use a pointer.
- Compare references with `Equal`, not `==` or string comparison against
  `String()`, so callers don't have to think about canonicalization
  themselves — though in practice `Equal` just compares the canonical
  string, so both work identically for an already-`Parse`d value.

## Representing a reference field in JSON Schema

Every JSON Schema your IR ships should describe a reference field as a
canonical-URI string, not as an unconstrained string. Two ways to get
there, depending on how your schema is produced.

### Hand-authored or `$ref`-composed schemas

Point at this module's published fragment instead of restating
`{"type": "string", "format": "uri"}` in every schema that has a
reference field:

```json
{
  "properties": {
    "sourceRepo": {
      "$ref": "https://raw.githubusercontent.com/plexusone/resourceref/main/schema/resourceref.schema.json"
    }
  }
}
```

The fragment is also embedded in this module as
[`schema.ResourceRefSchema`](../schema/embed.go), for generators that want
to vendor it locally (via `//go:embed`) instead of resolving `$ref` over
the network at lint/build time.

### Go-first schemas generated with invopop/jsonschema

If your repo generates its own JSON Schema from Go types the way this
module does (`invopop/jsonschema`, per the org's Go-first convention),
reflection alone won't render a `resourceref.ResourceRef` field usefully —
the reflector will try to introspect its unexported internal field.
`ResourceRef` deliberately does **not** implement invopop's
`JSONSchema() *jsonschema.Schema` hook itself, because doing so would
require this module to import `invopop/jsonschema` at runtime, breaking
its stdlib-only constraint. Add the mapping in your own generator instead,
using `Reflector.Mapper`:

```go
import (
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/plexusone/resourceref"
)

var resourceRefType = reflect.TypeOf(resourceref.ResourceRef{})

r := new(jsonschema.Reflector)
r.Mapper = func(t reflect.Type) *jsonschema.Schema {
	if t == resourceRefType {
		return &jsonschema.Schema{
			Type:        "string",
			Format:      "uri",
			Description: "A resourceref.ResourceRef, encoded as its canonical URI string.",
		}
	}
	return nil
}
s := r.Reflect(&YourIRType{})
```

This is the same technique this module's own `gen/main.go` would use if a
future type in this module needed to embed another module's opaque type —
verified working (a `ResourceRef` field renders as
`{"type": "string", "format": "uri"}`, not as an object with a `raw`
property) before being written up here.

## Logical identity vs. instance identity

The most common mistake when adding references to an IR is making one
identifier mean two different things depending on context. Don't.
Separate them into distinct resource types, connected by a relationship
field:

```text
prn://acme/billing/org_7f3/invoice/inv_9c3        the conceptual invoice
prn://acme/billing/org_7f3/invoice-run/run_abc123 one specific computation of it
https://github.com/acme/billing-service            its source repository
```

A build artifact referencing "the invoice that produced me" should carry
an `invoiceRunPrn` (the specific run — reproducible, immutable), not an
`invoicePrn` (the conceptual invoice, which has no fixed inputs/outputs on
its own). A deployment status referencing "what code is running" should
point at a repository URL or a versioned artifact reference, not at a
mutable "current service" identifier that means something different every
time someone looks at it.

Field-naming convention: suffix reference fields with what kind of
resource they point to, not with generic terms like `Ref` or `ID` alone —
`sourceRepo`, `invoiceRunPrn`, `deployedServicePrn`. A reader (or an
agent) should be able to tell what's on the other end of a reference field
from its name alone, before resolving it.

## When you need more than a reference

`ResourceRef` alone is enough for lineage, audit trails, and "what does
this point to" fields — anywhere a reference is read but never
structurally validated or acted on. If your IR needs to *constrain* which
references are valid (only your own organization's resources, only
specific resource types, only versioned/pinned references) or *decompose* a
reference into typed fields (which org, which app, which resource type),
that's what the `scheme` subpackage's `Profile` format is for — see
[`scheme/profile.go`](../scheme/profile.go) and the reference profile in
[`scheme/testdata/reference-profile.json`](../scheme/testdata/reference-profile.json)
for a worked example. Don't hand-roll segment-splitting logic against
`ResourceRef.String()` in your own repo; that is exactly the mistake a
shared `Profile` engine exists to prevent.
