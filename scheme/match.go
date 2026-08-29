package scheme

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/plexusone/resourceref"
)

// ErrNoMatch is the sentinel wrapped by every error Validate and Split
// return because a reference does not conform to the profile — as opposed
// to the profile document itself being malformed (see ErrInvalidProfile,
// which a well-formed [*Profile] obtained from Load can never trigger
// here; it can still occur if a Profile literal was hand-constructed
// with an invalid regular expression, bypassing Load's checks).
var ErrNoMatch = errors.New("scheme: reference does not match profile")

// Parts is a reference's typed decomposition against the [Shape] it
// matched. Fields the matched shape's segments do not use are left at
// their zero value.
type Parts struct {
	Authority    string
	App          string
	Org          string
	ResourceType string
	ID           string
	Version      string
}

// Validate reports whether ref conforms to p: the correct scheme, an
// allowed set of URI components, a registered authority, a path matching
// one of p's Shapes (tried in declared order), a version/alias suffix
// satisfying p's VersionGrammar, and — where p registers resource types
// for ref's (authority, app) pair — a registered resourceType.
func (p *Profile) Validate(ref resourceref.ResourceRef) error {
	_, err := p.match(ref)
	return err
}

// Split validates ref against p and returns its typed decomposition.
// Callers built on top of a profile should always use Split rather than
// re-parsing a reference's raw string themselves — that is the entire
// reason a profile exists.
func (p *Profile) Split(ref resourceref.ResourceRef) (Parts, error) {
	return p.match(ref)
}

func (p *Profile) match(ref resourceref.ResourceRef) (Parts, error) {
	if ref.IsZero() {
		return Parts{}, fmt.Errorf("%w: zero ResourceRef", ErrNoMatch)
	}
	if !strings.EqualFold(ref.Scheme(), p.Scheme) {
		return Parts{}, fmt.Errorf("%w: scheme %q does not match profile scheme %q", ErrNoMatch, ref.Scheme(), p.Scheme)
	}

	// ref.String() is always ResourceRef's canonical form — an invariant
	// Parse guarantees — so this reparse cannot fail for any validly
	// constructed ResourceRef. Re-deriving structure here, rather than
	// exposing Host()/PathSegments() on ResourceRef itself, keeps that
	// type's public surface small and stable: the profile engine is the
	// one thing in this module that needs URI structure, so it owns
	// parsing it.
	u, err := url.Parse(ref.String())
	if err != nil {
		return Parts{}, fmt.Errorf("%w: canonical form failed to reparse: %w", ErrNoMatch, err)
	}

	if err := p.checkURISubset(u); err != nil {
		return Parts{}, err
	}

	authority, err := p.matchAuthority(u)
	if err != nil {
		return Parts{}, err
	}

	segments := pathSegments(u)
	version, segments, err := p.extractVersion(segments)
	if err != nil {
		return Parts{}, err
	}

	if err := p.checkSegmentPattern(authority, segments); err != nil {
		return Parts{}, err
	}

	parts, err := p.matchShape(segments)
	if err != nil {
		return Parts{}, err
	}
	parts.Authority = authority
	parts.Version = version

	if err := p.checkResourceType(parts); err != nil {
		return Parts{}, err
	}

	return parts, nil
}

func (p *Profile) checkURISubset(u *url.URL) error {
	if !p.URISubset.AllowQuery && (u.RawQuery != "" || u.ForceQuery) {
		return fmt.Errorf("%w: query component is not permitted by this scheme", ErrNoMatch)
	}
	if !p.URISubset.AllowFragment && (u.Fragment != "" || u.RawFragment != "") {
		return fmt.Errorf("%w: fragment component is not permitted by this scheme", ErrNoMatch)
	}
	if !p.URISubset.AllowUserinfo && u.User != nil {
		return fmt.Errorf("%w: userinfo component is not permitted by this scheme", ErrNoMatch)
	}
	if !p.URISubset.AllowPort && u.Port() != "" {
		return fmt.Errorf("%w: port component is not permitted by this scheme", ErrNoMatch)
	}
	return nil
}

func (p *Profile) matchAuthority(u *url.URL) (string, error) {
	if u.Opaque != "" || u.OmitHost || u.Host == "" {
		return "", fmt.Errorf("%w: reference has no authority", ErrNoMatch)
	}
	for _, a := range p.Authorities {
		if strings.EqualFold(a.Name, u.Host) {
			return a.Name, nil
		}
	}
	return "", fmt.Errorf("%w: authority %q is not registered", ErrNoMatch, u.Host)
}

// pathSegments splits a hierarchical reference's path on "/", using the
// escaped (still percent-encoded) form so a reserved character like an
// encoded "/" (%2F) within a segment can never be mistaken for a
// path-segment boundary.
func pathSegments(u *url.URL) []string {
	path := strings.TrimPrefix(u.EscapedPath(), "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

// extractVersion looks for p's version/alias suffix on the final path
// segment, classifies it against VersionPattern/AliasPattern, and returns
// the remaining segments with that suffix removed from the last one.
func (p *Profile) extractVersion(segments []string) (version string, remaining []string, err error) {
	if p.Version.Separator == "" {
		return "", segments, nil
	}
	if len(segments) == 0 {
		if p.Version.Required {
			return "", segments, fmt.Errorf("%w: version or alias suffix is required", ErrNoMatch)
		}
		return "", segments, nil
	}

	last := segments[len(segments)-1]
	idx := strings.LastIndex(last, p.Version.Separator)
	if idx == -1 {
		if p.Version.Required {
			return "", segments, fmt.Errorf("%w: version or alias suffix is required", ErrNoMatch)
		}
		return "", segments, nil
	}

	idPart, verPart := last[:idx], last[idx+len(p.Version.Separator):]
	if idPart == "" {
		return "", segments, fmt.Errorf("%w: empty segment before version suffix", ErrNoMatch)
	}
	if verPart == "" {
		return "", segments, fmt.Errorf("%w: empty version/alias suffix", ErrNoMatch)
	}

	matched, err := p.versionOrAliasMatches(verPart)
	if err != nil {
		return "", segments, err
	}
	if !matched {
		return "", segments, fmt.Errorf("%w: suffix %q matches neither versionPattern nor aliasPattern", ErrNoMatch, verPart)
	}

	out := append([]string(nil), segments...)
	out[len(out)-1] = idPart
	return verPart, out, nil
}

func (p *Profile) versionOrAliasMatches(s string) (bool, error) {
	if p.Version.VersionPattern != "" {
		re, err := regexp.Compile(p.Version.VersionPattern)
		if err != nil {
			return false, fmt.Errorf("%w: versionPattern: %w", ErrInvalidProfile, err)
		}
		if re.MatchString(s) {
			return true, nil
		}
	}
	if p.Version.AliasPattern != "" {
		re, err := regexp.Compile(p.Version.AliasPattern)
		if err != nil {
			return false, fmt.Errorf("%w: aliasPattern: %w", ErrInvalidProfile, err)
		}
		if re.MatchString(s) {
			return true, nil
		}
	}
	return false, nil
}

func (p *Profile) checkSegmentPattern(authority string, segments []string) error {
	if p.URISubset.SegmentPattern == "" {
		return nil
	}
	re, err := regexp.Compile(p.URISubset.SegmentPattern)
	if err != nil {
		return fmt.Errorf("%w: uriSubset.segmentPattern: %w", ErrInvalidProfile, err)
	}
	if !re.MatchString(authority) {
		return fmt.Errorf("%w: authority %q does not match segmentPattern", ErrNoMatch, authority)
	}
	for i, seg := range segments {
		if !re.MatchString(seg) {
			return fmt.Errorf("%w: path segment %d (%q) does not match segmentPattern", ErrNoMatch, i, seg)
		}
	}
	return nil
}

// matchShape tries p.Shapes in declared order and returns the typed Parts
// of the first one whose segment structure fits segments: an exact count
// match for a shape with no rest segment, or at-least that many for a
// shape whose final segment captures the rest.
func (p *Profile) matchShape(segments []string) (Parts, error) {
	for _, shape := range p.Shapes {
		n := len(shape.Segments)
		restIndex := -1
		for i, seg := range shape.Segments {
			if seg.Rest {
				restIndex = i
			}
		}
		if restIndex == -1 {
			if len(segments) != n {
				continue
			}
		} else if len(segments) < n {
			continue
		}

		var parts Parts
		for i, seg := range shape.Segments {
			value := segments[i]
			if seg.Rest {
				value = strings.Join(segments[i:], "/")
			}
			switch seg.Role {
			case RoleApp:
				parts.App = value
			case RoleOrg:
				parts.Org = value
			case RoleResourceType:
				parts.ResourceType = value
			case RoleID:
				parts.ID = value
			}
		}
		return parts, nil
	}
	return Parts{}, fmt.Errorf("%w: no shape matches a path with %d segment(s)", ErrNoMatch, len(segments))
}

// checkResourceType enforces p.ResourceTypes when — and only when — an
// entry is registered for parts' (Authority, App) pair. No entry means no
// restriction: the registry is opt-in and additive, per [Profile]'s
// ResourceTypes documentation.
func (p *Profile) checkResourceType(parts Parts) error {
	if len(p.ResourceTypes) == 0 {
		return nil
	}
	var candidates []string
	for _, rt := range p.ResourceTypes {
		if !strings.EqualFold(rt.Authority, parts.Authority) || !strings.EqualFold(rt.App, parts.App) {
			continue
		}
		candidates = append(candidates, rt.Types...)
	}
	if candidates == nil {
		return nil
	}
	for _, t := range candidates {
		if t == parts.ResourceType {
			return nil
		}
	}
	return fmt.Errorf("%w: resourceType %q is not registered for authority %q app %q",
		ErrNoMatch, parts.ResourceType, parts.Authority, parts.App)
}
