// Package scheme defines owned URI schemes as data.
//
// A Profile describes a namespace owner's "owned" resource-reference
// scheme as a SPIFFE-strict URI subset (no query, fragment, userinfo, or
// port), a registry of authorities (namespace owners), per-authority path
// shapes, resource-type registries, and version/alias suffix grammar.
//
// Profiles are loaded from JSON and validated against a published JSON
// Schema; this package supplies the engine that validates a
// [resourceref.ResourceRef] against a Profile and splits it into typed
// parts. It does not ship any specific scheme's profile data — that is
// defined by whoever adopts this package and loaded at runtime or
// embedded at build time.
package scheme
