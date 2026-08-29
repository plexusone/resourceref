package scheme

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// ErrInvalidProfile is the sentinel wrapped by every error Load returns
// because the profile document itself is structurally invalid. It does
// not cover reference-validation failures (a valid profile rejecting a
// particular [resourceref.ResourceRef]) — that is [Profile.Validate]'s
// concern.
var ErrInvalidProfile = errors.New("scheme: invalid profile")

// schemeNamePattern is the RFC 3986 scheme grammar: ALPHA *( ALPHA / DIGIT
// / "+" / "-" / "." ).
var schemeNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*$`)

// Load parses and structurally validates a Profile document.
//
// Load checks that the document is internally consistent — required
// fields are present, shapes are well-formed, every regular expression
// compiles, and every cross-reference within the document (a
// ResourceTypeRegistry entry's Authority, for example) resolves — but it
// does not validate the profile against the generated JSON Schema at
// runtime: this package is stdlib-only, and the schema exists for
// external tooling (schemakit, documentation, other languages), not as a
// second, dependency-bearing validation path here.
func Load(data []byte) (*Profile, error) {
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProfile, err)
	}
	if err := validate(&p); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProfile, err)
	}
	return &p, nil
}

func validate(p *Profile) error {
	if !schemeNamePattern.MatchString(p.Scheme) {
		return fmt.Errorf("scheme %q does not match the URI scheme grammar", p.Scheme)
	}

	if err := validateURISubset(p.URISubset); err != nil {
		return err
	}

	if len(p.Authorities) == 0 {
		return errors.New("at least one authority is required")
	}
	authorityNames := make(map[string]bool, len(p.Authorities))
	for i, a := range p.Authorities {
		if a.Name == "" {
			return fmt.Errorf("authorities[%d]: name is required", i)
		}
		if authorityNames[a.Name] {
			return fmt.Errorf("authorities[%d]: duplicate authority name %q", i, a.Name)
		}
		authorityNames[a.Name] = true
	}

	if len(p.Shapes) == 0 {
		return errors.New("at least one shape is required")
	}
	for i, s := range p.Shapes {
		if err := validateShape(s); err != nil {
			return fmt.Errorf("shapes[%d] (%s): %w", i, s.Name, err)
		}
	}

	for i, rt := range p.ResourceTypes {
		if err := validateResourceTypeRegistry(rt, authorityNames); err != nil {
			return fmt.Errorf("resourceTypes[%d]: %w", i, err)
		}
	}

	if err := validateVersionGrammar(p.Version); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	return nil
}

func validateURISubset(u URISubset) error {
	if u.SegmentPattern == "" {
		return nil
	}
	if _, err := regexp.Compile(u.SegmentPattern); err != nil {
		return fmt.Errorf("uriSubset.segmentPattern: invalid regular expression: %w", err)
	}
	return nil
}

func validateShape(s Shape) error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if len(s.Segments) == 0 {
		return errors.New("at least one segment is required")
	}

	restIndex := -1
	seenRoles := make(map[string]bool, len(s.Segments))
	for i, seg := range s.Segments {
		switch seg.Role {
		case RoleApp, RoleOrg, RoleResourceType, RoleID:
			// valid role
		default:
			return fmt.Errorf("segments[%d]: unknown role %q", i, seg.Role)
		}
		if seenRoles[seg.Role] {
			// A shape must not use the same role twice: Split assigns one
			// Parts field per role, so a second occurrence would silently
			// overwrite the first rather than surface as a mismatch.
			return fmt.Errorf("segments[%d]: role %q already used earlier in this shape", i, seg.Role)
		}
		seenRoles[seg.Role] = true
		if seg.Rest {
			if seg.Role != RoleID {
				return fmt.Errorf("segments[%d]: only an %q segment may set rest", i, RoleID)
			}
			if restIndex != -1 {
				return errors.New("at most one segment may set rest")
			}
			restIndex = i
		}
	}
	if restIndex != -1 && restIndex != len(s.Segments)-1 {
		return errors.New("the rest segment must be the last segment")
	}

	hasResourceType := false
	for _, seg := range s.Segments {
		if seg.Role == RoleResourceType {
			hasResourceType = true
			break
		}
	}
	if !hasResourceType {
		return fmt.Errorf("every shape must include a %q segment", RoleResourceType)
	}

	return nil
}

func validateResourceTypeRegistry(rt ResourceTypeRegistry, authorityNames map[string]bool) error {
	if rt.Authority == "" {
		return errors.New("authority is required")
	}
	if !authorityNames[rt.Authority] {
		return fmt.Errorf("authority %q is not a registered authority", rt.Authority)
	}
	if len(rt.Types) == 0 {
		return errors.New("at least one type is required")
	}
	return nil
}

func validateVersionGrammar(v VersionGrammar) error {
	needsPattern := v.Required || v.Separator != ""
	hasPattern := v.VersionPattern != "" || v.AliasPattern != ""
	if needsPattern && !hasPattern {
		return errors.New("versionPattern or aliasPattern is required when separator is set or version is required")
	}
	if v.VersionPattern != "" {
		if _, err := regexp.Compile(v.VersionPattern); err != nil {
			return fmt.Errorf("versionPattern: invalid regular expression: %w", err)
		}
	}
	if v.AliasPattern != "" {
		if _, err := regexp.Compile(v.AliasPattern); err != nil {
			return fmt.Errorf("aliasPattern: invalid regular expression: %w", err)
		}
	}
	return nil
}
