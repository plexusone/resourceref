# PRD — resourceref — Universal Resource References & the Scheme Engine

**Initiative:** `INIT-RESOURCEREF-001`
**Repository:** `github.com/plexusone/resourceref`

## Problem

Organizations running multiple independently developed services and JSON
IRs need to refer to the same things — repos, services, agents, workflow
runs, tickets, datasets, findings. Without a shared reference mechanism,
each system invents its own `resource_id` convention, and cross-system
joins (policy, audit, lineage, cost, graphs) require pairwise translation
between every pair of systems that needs to talk about the same resource.

AWS demonstrates the alternative at large scale: one canonical
resource-reference format (the ARN) becomes the foreign key for an entire
platform, letting independently developed services (IAM, CloudTrail,
Config, cost reporting, security tooling) refer to the same resource
without a shared schema. No general-purpose, standards-based equivalent
exists as a reusable library for organizations that want the same
property without adopting AWS's specific, service-coupled syntax.

## Product

`resourceref` is a small, dependency-light Go module providing the
*mechanism* for universal resource references; namespace owners define
their own *instances* on top of it:

- **ResourceRef** — a resource reference is a URI. The universal rule:
  prefer the resource owner's canonical URI when one exists (`arn:aws:...`,
  `spiffe://...`, `https://github.com/...`); mint an owned URI only when one
  does not. External resources need no translation layer — the ARN *is*
  the reference.
- **Scheme profiles** — data-driven definitions of owned URI schemes:
  a SPIFFE-strict URI subset (no query/fragment/userinfo/port, strict
  casing), authority registry, per-authority segment shapes, resource-type
  registries, and version/alias suffix grammar (`@prod`, `@14`).
- **Tooling** — parse, validate, canonicalize, compare; JSON/Text
  marshaling; JSON Schema for embedding refs in IRs (Go-first generated).

This module ships no scheme instance of its own, only a reference test
fixture. Any owned scheme (its authorities, resource types, path shapes)
is defined entirely by whoever adopts the engine.

## Users

- **Spec/IR authors** — one `ref` field convention instead of a bespoke ID
  string per document format.
- **Invocation/interop layers** built on top — resource references as
  typed, versionable invocation targets.
- **Multi-service platforms** — reference fields in API schemas, run
  lineage, authorization resource mapping.
- **Graph/reporting systems** — refs as node IDs and join keys across
  independently developed producers.

## Key requirements

1. Standards-based: RFC 3986 syntax, RFC 8141 persistence semantics, SPIFFE
   strictness lessons, ARN cross-referencing lessons. Do not invent
   low-level syntax.
2. Identifier ≠ locator: no resolution in this module (resolvers live
   above); no DNS coupling in canonical identity.
3. Logical vs instance identity are distinct resource types, related by
   edges — never one identifier meaning repo+service+deployment at once.
4. Scheme profiles are data (JSON, schema-validated), so instances can
   live anywhere and third parties can define their own schemes.
5. Stdlib-only Go; tiny; stable — a leaf module safe to depend on widely
   without dragging in a heavy dependency tree.

## Non-goals

- Resolution, registries-as-services, HTTP endpoints (`https://id.…`
  resolvable IDs are a resolver-layer concern, deliberately excluded).
- Invocation, authorization, run semantics — those belong to a layer
  built on top of this module.
- Defining any specific scheme instance — that is always the adopter's
  responsibility, never this module's.

## Success criteria

- A downstream invocation/interop layer can adopt this module's PRN
  handling with no local grammar code of its own.
- An external-URI passthrough (ARN, SPIFFE ID, GitHub URL) round-trips
  ResourceRef marshaling untouched.
- A second scheme instance, defined by an unrelated adopter, is definable
  purely as profile data, with no changes to this module.
