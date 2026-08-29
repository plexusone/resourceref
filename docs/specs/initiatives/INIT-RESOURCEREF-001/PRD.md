# PRD — resourceref — Universal Resource References & the PRN Scheme Engine

**Initiative:** `INIT-RESOURCEREF-001`
**Repository:** `github.com/plexusone/resourceref`

## Problem

Multiple JSON IRs and services across the PlexusOne / ProductBuildersHQ /
grokify ecosystem need to refer to the same things — repos, services, agents,
notebooks, workflows, runs, roadmap items, threat models, findings, policies,
datasets. Without a shared reference mechanism, each spec invents its own
`resource_id` convention and cross-system joins (policy, audit, lineage, cost,
graphs) require pairwise translation. AWS demonstrates the alternative: one
canonical resource-reference format becomes the foreign key for the entire
platform.

No such mechanism exists in the ecosystem today (verified: no URN/ARN scheme
anywhere in omniworkflow/omniskill/omniagent or the Forge repos).

## Product

`resourceref` is an org-neutral spec + Go module — the design-system-spec /
api-style-spec pattern applied to resource naming. It provides the
*mechanism*; namespace owners define *instances*:

- **ResourceRef** — a resource reference is a URI. The universal rule:
  prefer the resource owner's canonical URI when one exists (`arn:aws:...`,
  `spiffe://...`, `https://github.com/...`); mint an owned URI only when one
  does not. External resources need no translation layer — the ARN *is* the
  reference.
- **Scheme profiles** — data-driven definitions of owned URI schemes:
  a SPIFFE-strict URI subset (no query/fragment/userinfo/port, strict
  casing), authority registry, per-authority segment shapes, resource-type
  registries, and version/alias suffix grammar (`@prod`, `@14`).
- **Tooling** — parse, validate, canonicalize, compare; JSON/Text
  marshaling; JSON Schema for embedding refs in IRs (Go-first generated).

PlexusOne's instance — the `prn://` scheme (authorities, per-app resource
types, tenancy profile) — is defined in `plexusone-standards`, not here.
CrossForge consumes both and adds invocation semantics on top.

## Users

- **Spec/IR authors** (prism/PBHQ artifacts, threat models, signal-spec,
  api-style-spec) — one `ref` field convention instead of bespoke IDs.
- **CrossForge** — PRNs as invocation targets; ResourceRef in lineage.
- **Forge products** — PRN fields in API schemas (`workflowPrn`), run
  lineage, authz resource mapping.
- **Graph/reporting systems** — refs as node IDs and join keys across
  independently developed producers.

## Key requirements

1. Standards-based: RFC 3986 syntax, RFC 8141 persistence semantics, SPIFFE
   strictness lessons, ARN cross-referencing lessons. Do not invent
   low-level syntax.
2. Identifier ≠ locator: no resolution in this module (resolvers live above,
   e.g. crossforge/systemforge); no DNS coupling in canonical identity.
3. Logical vs instance identity are distinct resource types, related by
   edges — never one identifier meaning repo+service+deployment at once.
4. Scheme profiles are data (JSON, schema-validated), so instances live in
   standards repos and third parties can define their own schemes.
5. Stdlib-only Go; tiny; stable — this becomes the most-imported leaf module
   in the ecosystem.

## Non-goals

- Resolution, registries-as-services, HTTP endpoints (`https://id.…`
  resolvable IDs are a resolver-layer concern, deliberately excluded).
- Invocation, authorization, run semantics (crossforge).
- Defining the PlexusOne PRN instance itself (plexusone-standards).

## Success criteria

- CrossForge's PRN handling is an import of this module plus the
  plexusone-standards profile, with no local grammar code.
- An external-URI passthrough (ARN, SPIFFE ID, GitHub URL) round-trips
  ResourceRef marshaling untouched.
- A second scheme instance (e.g. a ProductBuildersHQ authority for roadmap
  artifacts) is definable purely as profile data, no Go changes.
