// Package resourceref provides a universal resource-reference type and a
// data-driven engine for defining owned URI schemes.
//
// The governing rule is simple: a resource reference is a URI. Prefer the
// resource owner's canonical URI when one exists — an AWS ARN, a SPIFFE ID,
// a GitHub URL — and mint an owned URI only when one does not. Owner-canonical
// URIs pass through this package untouched; there is no translation layer for
// external resources.
//
// # Two layers
//
// ResourceRef is the universal type: any RFC 3986 URI, canonicalized and
// comparable, usable as a graph node ID, a lineage edge, or a field in any
// JSON IR.
//
// The scheme subpackage defines owned URI schemes as data (a Profile), not
// code: a SPIFFE-strict URI subset, an authority (namespace owner) registry,
// per-authority path shapes, resource-type registries, and version/alias
// suffix grammar. Namespace owners define their own profiles; this package
// supplies only the mechanism, not any specific scheme instance.
//
// # What this package deliberately does not do
//
// It does not resolve references to locations — identifier is not locator.
// It does not authorize, invoke, or execute anything against a reference.
// It does not define any specific scheme instance. Those concerns belong
// to resolvers, invocation contracts, and whatever adopts this module.
package resourceref
