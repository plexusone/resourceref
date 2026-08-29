# resourceref

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/resourceref/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/resourceref/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/resourceref/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/resourceref/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/resourceref/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/resourceref/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/resourceref
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/resourceref
 [repo-url]: https://github.com/plexusone/resourceref
 [license-svg]: https://img.shields.io/badge/license-Apache%202.0-blue.svg
 [license-url]: https://github.com/plexusone/resourceref/blob/main/LICENSE

Universal resource references and a data-driven engine for owned URI
schemes, for Go.

## The idea

A resource reference is a URI. Prefer the resource owner's canonical URI
when one exists — an AWS ARN, a SPIFFE ID, a GitHub URL — and mint an owned
URI only when one does not. Owner-canonical URIs pass through this module
untouched; there is no translation layer for external resources.

```text
arn:aws:s3:::customer-data                  (AWS's own reference — pass through)
spiffe://prod.example.com/ns/api/sa/default  (SPIFFE's own reference — pass through)
https://github.com/acme/billing-service      (GitHub's own reference — pass through)
prn://acme/billing/org_7f3/invoice/inv_9c3@prod   (an owned reference)
```

Once independently developed systems agree on this one reference format,
they can exchange references — in JSON IRs, policy documents, lineage
graphs, audit events — without understanding one another's internal
schemas. The reference becomes the join key.

## What's in this module

- **`ResourceRef`** — a parsed, canonicalized, comparable resource
  reference. `Parse`, `String`, `Equal`, JSON/text marshaling as a plain
  string.
- **`scheme`** — the engine for defining *owned* URI schemes as data: a
  `Profile` describes a SPIFFE-strict URI subset, an authority (namespace
  owner) registry, per-authority path shapes, resource-type registries,
  and version/alias suffix grammar (`@prod`, `@14`). `Load` parses and
  structurally validates a profile document; `Profile.Validate` and
  `Profile.Split` check a `ResourceRef` against it and, for `Split`,
  return its typed decomposition (`Parts`). Namespace owners publish their
  own profile — this module ships none, only a reference test fixture.
- **`schema`** — the generated JSON Schema for the `Profile` format and a
  ready-to-embed fragment for a `ResourceRef`-typed field, for external
  tooling and other IRs' schemas.

See [docs/embedding.md](docs/embedding.md) for how to add a reference
field to your own IR — Go type, JSON Schema, and the logical-vs-instance
identity pitfall to avoid.

## What this module deliberately does not do

- **No resolution.** Identifier is not locator; turning a reference into an
  endpoint is a resolver's job, not this module's.
- **No specific scheme.** This module ships the mechanism only — a
  reference test fixture, not a real scheme. Defining your own owned
  scheme (its authorities, path shapes, resource types) is entirely up to
  whoever adopts it; see [docs/embedding.md](docs/embedding.md).
- **No invocation or authorization.** Calling, authorizing, or executing
  against a reference is the concern of a layer built on top of this
  module, not this module's own.

## Design basis

RFC 3986 (URI generic syntax) for the wire form, RFC 8141 (URN persistence
semantics) for the identifier-not-locator discipline, and the SPIFFE ID
specification for strict URI-subset validation. AWS ARN and Kubernetes
resource references inform the cross-referencing and typed-identity lessons
without adopting their service-specific syntax.

## Status

Early development. See
[`docs/specs/initiatives/INIT-RESOURCEREF-001`](docs/specs/initiatives/INIT-RESOURCEREF-001)
for the current PRD/TRD/PLAN/ROADMAP.

## License

Apache License 2.0. See [LICENSE](LICENSE).
