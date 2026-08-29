package resourceref

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrNoScheme is returned by Parse when the input has no URI scheme. A
// ResourceRef must be an absolute URI reference — relative references are
// not resource references.
var ErrNoScheme = errors.New("resourceref: missing URI scheme")

// A ResourceRef is a parsed, canonicalized resource reference.
//
// Construct one with Parse or MustParse; the zero value is not a valid
// reference (its String method returns "").
type ResourceRef struct {
	raw string
}

// Parse parses s as an absolute URI reference (RFC 3986) and returns its
// canonical ResourceRef form.
//
// Canonicalization is conservative and never semantically reinterprets a
// reference's components:
//
//   - the scheme, and the host of a hierarchical reference, are lowercased;
//   - percent-encoding is normalized per RFC 3986 §6.2.2.2 — an encoded
//     octet is decoded only when it corresponds to an unreserved character
//     (ALPHA / DIGIT / "-" / "." / "_" / "~"), and any encoding that
//     remains has its hex digits uppercased;
//   - dot segments are removed from hierarchical paths per RFC 3986 §5.2.4;
//   - no scheme-specific default (e.g. treating :443 as implied by https)
//     is ever stripped or assumed.
//
// Owner-canonical external references — an AWS ARN, a SPIFFE ID, a URN —
// pass through with their opaque or authority structure intact.
func Parse(s string) (ResourceRef, error) {
	if s == "" {
		return ResourceRef{}, fmt.Errorf("resourceref: parse %q: %w", s, ErrNoScheme)
	}
	u, err := url.Parse(s)
	if err != nil {
		return ResourceRef{}, fmt.Errorf("resourceref: parse %q: %w", s, err)
	}
	if u.Scheme == "" {
		return ResourceRef{}, fmt.Errorf("resourceref: parse %q: %w", s, ErrNoScheme)
	}
	return ResourceRef{raw: canonicalize(u)}, nil
}

// MustParse is like Parse but panics if s cannot be parsed. It is intended
// for tests and package-level initialization from known-good literals, not
// for parsing untrusted input.
func MustParse(s string) ResourceRef {
	r, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return r
}

// String returns the reference's canonical form, or "" for the zero value.
func (r ResourceRef) String() string { return r.raw }

// IsZero reports whether r is the zero ResourceRef.
func (r ResourceRef) IsZero() bool { return r.raw == "" }

// Scheme returns the reference's URI scheme, already lowercased by Parse.
func (r ResourceRef) Scheme() string {
	i := strings.IndexByte(r.raw, ':')
	if i < 0 {
		return ""
	}
	return r.raw[:i]
}

// Equal reports whether r and o refer to the same canonical reference.
func (r ResourceRef) Equal(o ResourceRef) bool { return r.raw == o.raw }

// MarshalText implements encoding.TextMarshaler, encoding the reference as
// its canonical string form. json.Marshal uses this automatically for any
// struct field of type ResourceRef, so it encodes as a plain JSON string.
//
// Marshaling the zero value is an error: a ResourceRef field is either a
// real reference or should be represented as *ResourceRef (nil, with
// omitempty) — never as a silent empty string.
func (r ResourceRef) MarshalText() ([]byte, error) {
	if r.IsZero() {
		return nil, errors.New("resourceref: cannot marshal zero ResourceRef")
	}
	return []byte(r.raw), nil
}

// UnmarshalText implements encoding.TextUnmarshaler by parsing text with
// Parse.
func (r *ResourceRef) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}
