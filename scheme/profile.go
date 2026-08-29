package scheme

// Segment roles a [SegmentSpec] may declare. A shape assigns each of its
// fixed path segments one of these roles; [Parts] has one field per role.
const (
	// RoleApp identifies the owning application segment of a tenant-scoped
	// shape (for example "dashforge" in a PlexusOne PRN).
	RoleApp = "app"
	// RoleOrg identifies the tenancy segment of a tenant-scoped shape.
	RoleOrg = "org"
	// RoleResourceType identifies the resource-type segment, present in
	// every shape.
	RoleResourceType = "resourceType"
	// RoleID identifies the resource-identifier segment. On a "rest"
	// segment it captures one or more remaining path segments joined by
	// "/".
	RoleID = "id"
)

// A Profile describes an owned URI scheme as data: PlexusOne's "prn"
// scheme is one instance of this format, defined in plexusone-standards,
// not in this package.
//
// Profile is the Go-first source of truth for the generated JSON Schema
// embedded in the schema subpackage — field documentation here becomes
// schema descriptions.
type Profile struct {
	// Scheme is the URI scheme this profile governs, e.g. "prn". It must
	// match the RFC 3986 scheme grammar: a letter followed by letters,
	// digits, "+", "-", or ".".
	Scheme string `json:"scheme"`

	// URISubset constrains which URI components an owned reference under
	// this scheme may use.
	URISubset URISubset `json:"uriSubset"`

	// Authorities is the registry of namespace owners permitted as the
	// URI host component of a reference under this scheme.
	Authorities []Authority `json:"authorities"`

	// Shapes are the permitted path-segment structures, tried in order;
	// the first shape whose segment count matches a reference's path wins.
	Shapes []Shape `json:"shapes"`

	// ResourceTypes optionally registers known resource-type values per
	// authority (and, for tenant-scoped shapes, per app). A reference
	// whose (authority, app) pair has no registry entry here is not
	// restricted — the registry is opt-in and additive, not a closed list
	// that must be updated before a new type can be used.
	ResourceTypes []ResourceTypeRegistry `json:"resourceTypes,omitempty"`

	// Version defines the "@version" / "@alias" suffix grammar applied to
	// a reference's final path segment.
	Version VersionGrammar `json:"version"`
}

// URISubset constrains a scheme to a stricter subset of general URI syntax,
// following the SPIFFE ID specification's lead: an owned reference scheme
// should say plainly which components are meaningful, rather than
// inheriting the full generality of RFC 3986.
type URISubset struct {
	// AllowQuery permits a "?query" component on references under this
	// scheme.
	AllowQuery bool `json:"allowQuery"`
	// AllowFragment permits a "#fragment" component.
	AllowFragment bool `json:"allowFragment"`
	// AllowUserinfo permits a "user:pass@" component in the authority.
	AllowUserinfo bool `json:"allowUserinfo"`
	// AllowPort permits an explicit ":port" component in the authority.
	AllowPort bool `json:"allowPort"`
	// SegmentPattern is a regular expression every authority and
	// non-version-suffixed path segment must match. Empty means no
	// additional charset restriction beyond valid URI syntax.
	SegmentPattern string `json:"segmentPattern,omitempty"`
}

// An Authority is a registered namespace owner permitted as the host
// component of a reference under a [Profile]'s scheme, for example
// "plexusone" or "productbuildershq".
type Authority struct {
	// Name is the authority's URI host value.
	Name string `json:"name"`
	// Description documents who owns this namespace and what it is for.
	Description string `json:"description,omitempty"`
}

// A Shape is one permitted path-segment structure for references under a
// [Profile]'s scheme, for example a tenant-scoped shape
// ({app}/{org}/{resourceType}/{id}) or an ecosystem-artifact shape
// ({resourceType}/{id...}).
type Shape struct {
	// Name identifies the shape for diagnostics, e.g. "tenant-scoped".
	Name string `json:"name"`
	// Segments describes the path, in order, after the authority. At most
	// one segment may set Rest to true, and it must be the last element.
	Segments []SegmentSpec `json:"segments"`
}

// A SegmentSpec is one path segment of a [Shape].
type SegmentSpec struct {
	// Role is the meaning of this segment: one of RoleApp, RoleOrg,
	// RoleResourceType, or RoleID.
	Role string `json:"role"`
	// Rest marks this as the shape's final segment, capturing one or more
	// remaining path segments (joined by "/") rather than exactly one.
	// Only RoleID segments may set Rest.
	Rest bool `json:"rest,omitempty"`
}

// A ResourceTypeRegistry entry restricts the resourceType segment for
// references scoped to one authority (and, for tenant-scoped shapes, one
// app) to a known set of values.
type ResourceTypeRegistry struct {
	// Authority is the namespace owner this entry applies to; it must be
	// one of the [Profile]'s registered Authorities.
	Authority string `json:"authority"`
	// App scopes this entry to one application's resource types under a
	// tenant-scoped shape. Empty applies to shapes with no app segment
	// (ecosystem artifacts).
	App string `json:"app,omitempty"`
	// Types is the set of permitted resourceType segment values for this
	// (Authority, App) pair.
	Types []string `json:"types"`
}

// VersionGrammar defines the "@version" / "@alias" suffix grammar applied
// to a reference's final path segment, distinguishing an immutable version
// (e.g. "14") from a mutable alias (e.g. "prod") pointing at one.
type VersionGrammar struct {
	// Separator is the character (or short string) that introduces the
	// suffix, e.g. "@". Empty means this scheme has no version suffix
	// grammar at all.
	Separator string `json:"separator,omitempty"`
	// Required, when true, means every reference to an invocable resource
	// under this scheme must carry a version or alias suffix — consumers
	// (for example crossforge) enforce this for invocation targets, this
	// package only records the policy.
	Required bool `json:"required,omitempty"`
	// VersionPattern is a regular expression matching an immutable
	// version token.
	VersionPattern string `json:"versionPattern,omitempty"`
	// AliasPattern is a regular expression matching a mutable alias
	// token.
	AliasPattern string `json:"aliasPattern,omitempty"`
}
