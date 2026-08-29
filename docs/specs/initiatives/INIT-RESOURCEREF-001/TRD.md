# TRD — resourceref — Universal Resource References & the PRN Scheme Engine

**Initiative:** `INIT-RESOURCEREF-001`
**Repository:** `github.com/plexusone/resourceref`

## Module constraints

- Module `github.com/plexusone/resourceref`, Go 1.26+, **stdlib-only**
  (generator deps behind `//go:build ignore` + `tools.go`).
- Leaf of the ecosystem dependency graph: imports nothing from plexusone /
  grokify / ProductBuildersHQ; everything may import it (specs, crossforge,
  forges). Never imports crossforge or systemforge.
- License: permissive (align with the crossforge decision — Apache-2
  preferred for the patent grant on a spec-like layer).

## Package layout

```text
resourceref/
├── resourceref.go   ResourceRef type, universal-URI handling, passthrough
├── scheme/          Profile type, profile loading/validation, scheme engine
├── schema/          JSON Schemas: ResourceRef embedding + profile format (//go:embed)
└── gen/             //go:build ignore schema generator + tools.go
```

## Core model

```go
// A resource reference is a URI. Owner-canonical URIs pass through
// untouched; owned schemes are validated against a Profile.
type ResourceRef struct {
    uri string // canonical form; unexported to force construction via Parse
}

func Parse(s string) (ResourceRef, error)          // RFC 3986 parse + canonicalize
func (r ResourceRef) String() string                // canonical serialization
func (r ResourceRef) Scheme() string
func (r ResourceRef) Equal(o ResourceRef) bool      // canonical comparison
// encoding.TextMarshaler/Unmarshaler + json round-trip as a plain string:
// {"ref": "prn://plexusone/dashforge/org_7f3/notebook/nb_x@prod"}
```

- Canonicalization: lowercase scheme/authority, no default-port trickery,
  percent-encoding normalized, no dot-segment paths. External URIs are
  canonicalized conservatively (never semantically rewritten).
- Comparison is on canonical form only; no network, no resolution.

## Scheme profiles (`scheme/`)

A Profile is data (JSON, schema-validated) defining an owned scheme:

```text
Profile
├── scheme            e.g. "prn"
├── uriSubset         SPIFFE-strict constraints: no query, fragment,
│                     userinfo, port; allowed charsets per segment
├── authorities[]     registered namespace owners (e.g. plexusone,
│                     productbuildershq, grokify)
├── shapes[]          per-authority path shapes, e.g.
│                     {app}/{org}/{resourceType}/{id}   (tenant-scoped)
│                     {resourceType}/{path...}          (ecosystem artifact)
├── resourceTypes     per-authority/app type registries (extensible)
└── version           suffix grammar: @{version|alias} on the final
                      segment; alias charset; whether version is required
```

```go
func Load(profile []byte) (*scheme.Profile, error)
func (p *Profile) Validate(r ResourceRef) error     // full profile check
func (p *Profile) Split(r ResourceRef) (Parts, error) // typed segment access
// Parts: Authority, App, Org, ResourceType, ID, Version — empty where the
// matched shape has no such segment.
```

- The engine matches a ref against the profile's shapes and returns typed
  parts; consumers (crossforge, forges) never re-parse strings.
- Profiles compose: an instance repo (plexusone-standards) ships the
  canonical `prn` profile JSON; consumers embed or vendor it.
- Version/alias semantics (immutable version vs mutable alias, "pin
  required" policy) are *expressed* in the profile; *enforced* by consumers
  (crossforge rejects unversioned invocation targets).

## Relationship to instances and consumers

```text
resourceref (mechanism: Go module + profile format)     ← this repo
    ↑ profile data
plexusone-standards/resourceref/prn.profile.json (instance: the prn scheme)
    ↑ imported/embedded by
crossforge (invocation targets MUST be profile-valid PRNs)
forges / specs / IRs (ResourceRef fields; PRN where owned)
```

Mirrors design-system-spec ↔ plexusone-standards/design-system and
api-style-spec ↔ plexusone-standards/api.

## Design rules carried from the ideation review

- Prefer owner-canonical URIs; mint owned URIs only when none exists.
- Identity is not location: no `github`/provider or DNS-resolvable hosts in
  owned canonical identities; resolution is a separate layer.
- Logical vs instance identity are separate resource types (service vs
  deployment vs repo), connected by relationship edges in consuming IRs —
  the profile's type registry keeps them distinct.

## Testing

- Golden corpus: valid/invalid refs per shape, canonicalization
  idempotence, external-URI passthrough fidelity (ARN, SPIFFE, https,
  urn:), profile round-trips, fuzz Parse.
- Schema tests: profile JSON Schema and ResourceRef embedding schema
  (schemakit-linted camelCase) validate shipped fixtures.
