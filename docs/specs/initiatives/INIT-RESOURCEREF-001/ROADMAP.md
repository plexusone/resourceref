# ROADMAP — resourceref — Universal Resource References & the Scheme Engine

**Initiative:** `INIT-RESOURCEREF-001`
**Repository:** `github.com/plexusone/resourceref`

A real scheme instance and a downstream invocation-layer consumer are
each their own separate project, tracked in their own roadmaps — neither
lives here, though both are what Phase 3 validates this module against.

## Phase 1 — ResourceRef Core

**Theme:** A resource reference is a URI — parse, canonicalize, compare, marshal

- [ ] `RMI-RESOURCEREF-001` Stdlib-only module scaffolding: go.mod, package layout, shared go-ci workflow, license decision
- [ ] `RMI-RESOURCEREF-002` ResourceRef type: RFC 3986 parse, conservative canonicalization, Equal, Text/JSON marshaling as plain string; external-URI passthrough corpus (arn:, spiffe://, https://, urn:) with fidelity invariants and fuzz tests
  - Depends on: `RMI-RESOURCEREF-001`

## Phase 2 — Scheme Profile Engine

**Theme:** Owned schemes as data — SPIFFE-strict subsets, shapes, typed parts, versions

- [ ] `RMI-RESOURCEREF-003` Profile format and loader: scheme, uriSubset constraints (no query/fragment/userinfo/port), authority registry, per-authority path shapes, resource-type registries, version/alias suffix grammar; JSON Schema generated Go-first
  - Depends on: `RMI-RESOURCEREF-002`
- [ ] `RMI-RESOURCEREF-004` Validation and shape matching: Profile.Validate, Profile.Split returning typed Parts (authority, app, org, resourceType, id, version); reference test profile and golden corpus
  - Depends on: `RMI-RESOURCEREF-003`

## Phase 3 — Instance & Consumer Stabilization

**Theme:** Proven by a real scheme instance and a real consumer before pinning a tag

- [ ] `RMI-RESOURCEREF-005` Embedding guidance and schemas for IR authors: ref-field JSON Schema fragment, a general embedding guide, logical-vs-instance identity guidance
  - Depends on: `RMI-RESOURCEREF-004`
- [ ] `RMI-RESOURCEREF-006` Stabilization: fold feedback from a real scheme instance and a real downstream consumer; tag v0.x with both building against the tag
  - Depends on: `RMI-RESOURCEREF-004`
