# PLAN — resourceref — Universal Resource References & the Scheme Engine

**Initiative:** `INIT-RESOURCEREF-001`
**Repository:** `github.com/plexusone/resourceref`

## Approach

Smallest possible leaf module, stabilized early: this is meant to be a
widely-importable dependency, so the priority is a tiny, correct, boring
API — one type, one profile format — proven against a real scheme
instance and a real invocation-layer consumer before any tag consumers
pin against.

Sequence mechanism → instance → consumer: the ResourceRef core and
profile engine land first; a real scheme instance is authored against it
separately (its own project, out of this module's scope); a downstream
invocation/interop layer then adopts the module instead of any local
grammar it might otherwise have built.

## Sequencing

1. **Phase 1 — ResourceRef core.** Module scaffolding; Parse/canonicalize/
   compare/marshal; external-URI passthrough corpus (ARN, SPIFFE, https,
   urn:). Pure functions, heavy table tests, fuzzing.
2. **Phase 2 — Scheme profile engine.** Profile format (JSON + generated
   schema), loader/validator, shape matching with typed Parts, version/alias
   suffix grammar. A reference test profile ships in-repo; any real scheme
   instance lives outside this module, wherever its owner keeps it.
3. **Phase 3 — Instance & consumer proof.** Validate the design against a
   real scheme instance and a real downstream consumer before declaring
   the API stable; fold feedback into a v0.x tag both depend on.

## Dependencies and coordination

- A real scheme instance (its authorities, shapes, resource types) is
  authored as a separate project against this module's Phase 2 profile
  format.
- A downstream invocation/interop layer adopts this module's profile
  engine rather than implementing its own reference grammar, and adds
  invocation-side helpers (require-a-reference, require-a-version) on top.
- License decision (permissive; Apache-2 preferred) is made once and
  shared with any sibling project that composes with this one.

## Risks

- **API churn in a widely-imported module.** Mitigation: unexported internal
  representation (construct via Parse only), tiny surface, no tag until
  real consumers have exercised it.
- **Profile format over-generalization.** Mitigation: the format supports
  exactly the shapes a real instance and one hypothetical second authority
  need; anything speculative is rejected.
- **Canonicalization subtleties on external URIs.** Mitigation:
  conservative normalization only; passthrough fidelity is a tested
  invariant, not best-effort.

## Definition of done

Phases 1–3 complete; a real scheme instance validates against the shipped
schema; a real downstream consumer adopts the module with no local
grammar code; fuzz + golden corpora green in CI; tagged v0.x with
consumers building against the tag.
