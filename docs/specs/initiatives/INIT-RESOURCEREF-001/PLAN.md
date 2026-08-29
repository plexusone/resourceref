# PLAN — resourceref — Universal Resource References & the PRN Scheme Engine

**Initiative:** `INIT-RESOURCEREF-001`
**Repository:** `github.com/plexusone/resourceref`

## Approach

Smallest possible leaf module, stabilized early: this becomes the
most-imported package in the ecosystem, so the priority is a tiny, correct,
boring API — one type, one profile format — proven against a real instance
(the PRN profile in plexusone-standards) and a real consumer (crossforge)
before any tag consumers pin.

Sequence mechanism → instance → consumer: the ResourceRef core and profile
engine land first; the PRN instance definition is authored in
plexusone-standards against it (that RMI lives in INIT-PLEXUSONESTANDARDS-001);
crossforge then swaps its planned local grammar for this module
(`RMI-CROSSFORGE-002` is re-scoped to consume resourceref + the PRN profile).

## Sequencing

1. **Phase 1 — ResourceRef core.** Module scaffolding; Parse/canonicalize/
   compare/marshal; external-URI passthrough corpus (ARN, SPIFFE, https,
   urn:). Pure functions, heavy table tests, fuzzing.
2. **Phase 2 — Scheme profile engine.** Profile format (JSON + generated
   schema), loader/validator, shape matching with typed Parts, version/alias
   suffix grammar. A reference test profile ships in-repo; the real PRN
   profile lives in plexusone-standards.
3. **Phase 3 — Instance & consumer proof.** Coordinate the PRN profile
   authoring in plexusone-standards and the crossforge adoption; fold
   feedback into a v0.x tag both depend on.

## Dependencies and coordination

- `INIT-PLEXUSONESTANDARDS-001` authors the PRN instance profile (added as
  an RMI there) — needs Phase 2's profile format.
- `INIT-CROSSFORGE-001` `RMI-CROSSFORGE-002` is re-scoped from "implement
  PRN grammar" to "adopt resourceref + PRN profile; add invocation-side
  helpers (require-PRN, require-version)".
- License decision shared with crossforge (permissive; Apache-2 preferred)
  — decide once for both before first tags.

## Risks

- **API churn in the most-imported module.** Mitigation: unexported internal
  representation (construct via Parse only), tiny surface, no tag until the
  crossforge + standards consumers have exercised it.
- **Profile format over-generalization.** Mitigation: the format supports
  exactly the shapes the PRN instance and one hypothetical second authority
  need; anything speculative is rejected.
- **Canonicalization subtleties on external URIs.** Mitigation:
  conservative normalization only; passthrough fidelity is a tested
  invariant, not best-effort.

## Definition of done

Phases 1–3 complete; the PRN profile in plexusone-standards validates against
the shipped schema; crossforge consumes the module with no local grammar
code; fuzz + golden corpora green in CI; tagged v0.x with both consumers
building against the tag.
