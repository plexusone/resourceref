package scheme

import (
	_ "embed"
	"errors"
	"testing"

	"github.com/plexusone/resourceref"
)

//go:embed testdata/reference-profile.json
var referenceProfileJSON []byte

func referenceProfile(t *testing.T) *Profile {
	t.Helper()
	p, err := Load(referenceProfileJSON)
	if err != nil {
		t.Fatalf("Load(reference profile fixture): %v", err)
	}
	return p
}

// TestSplitGoldenCorpus exercises Profile.Split against the reference
// profile fixture (scheme/testdata/reference-profile.json) with realistic
// references covering both shapes it defines: end-to-end proof that a
// profile authored against this engine produces the typed decomposition
// a real consumer needs, not just that the fixture parses.
func TestSplitGoldenCorpus(t *testing.T) {
	p := referenceProfile(t)

	cases := []struct {
		name string
		ref  string
		want Parts
	}{
		{
			name: "tenant-scoped with alias version",
			ref:  "prn://acme/billing/org_7f3/invoice/inv_9c3@prod",
			want: Parts{
				Authority: "acme", App: "billing", Org: "org_7f3",
				ResourceType: "invoice", ID: "inv_9c3", Version: "prod",
			},
		},
		{
			name: "tenant-scoped with numeric version",
			ref:  "prn://acme/shipping/org_7f3/shipment/ship_88a@14",
			want: Parts{
				Authority: "acme", App: "shipping", Org: "org_7f3",
				ResourceType: "shipment", ID: "ship_88a", Version: "14",
			},
		},
		{
			name: "tenant-scoped with no version (optional)",
			ref:  "prn://acme/shipping/org_7f3/shipment-run/run_abc",
			want: Parts{
				Authority: "acme", App: "shipping", Org: "org_7f3",
				ResourceType: "shipment-run", ID: "run_abc", Version: "",
			},
		},
		{
			name: "ecosystem artifact, single-segment id",
			ref:  "prn://opensource/roadmap/ISSUE-142",
			want: Parts{
				Authority: "opensource", ResourceType: "roadmap", ID: "ISSUE-142",
			},
		},
		{
			name: "ecosystem artifact, multi-segment rest id",
			ref:  "prn://opensource/roadmap/phase-1/ISSUE-142",
			want: Parts{
				Authority: "opensource", ResourceType: "roadmap", ID: "phase-1/ISSUE-142",
			},
		},
		{
			name: "app with no resourceType registry entry is unrestricted",
			ref:  "prn://acme/support/org_1/anything-goes/x",
			want: Parts{
				Authority: "acme", App: "support", Org: "org_1",
				ResourceType: "anything-goes", ID: "x",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ref := resourceref.MustParse(tc.ref)
			got, err := p.Split(ref)
			if err != nil {
				t.Fatalf("Split(%q): unexpected error: %v", tc.ref, err)
			}
			if got != tc.want {
				t.Fatalf("Split(%q) = %+v, want %+v", tc.ref, got, tc.want)
			}
			if err := p.Validate(ref); err != nil {
				t.Fatalf("Validate(%q): unexpected error: %v", tc.ref, err)
			}
		})
	}
}

func TestValidateRejectsGoldenCorpus(t *testing.T) {
	p := referenceProfile(t)

	cases := []struct {
		name string
		ref  string
	}{
		{
			name: "wrong scheme",
			ref:  "https://acme/billing/org_1/invoice/nb_1",
		},
		{
			name: "unregistered authority",
			ref:  "prn://unknown-authority/billing/org_1/invoice/nb_1",
		},
		{
			name: "resourceType not registered for this app",
			ref:  "prn://acme/billing/org_1/bogus-type/x",
		},
		{
			name: "version suffix matching neither pattern",
			ref:  "prn://acme/billing/org_1/invoice/nb_1@BAD!CHARS",
		},
		{
			name: "query component not permitted",
			ref:  "prn://acme/billing/org_1/invoice/nb_1?x=1",
		},
		{
			// The fixture's two shapes need exactly 4 segments
			// (tenant-scoped) or at least 2 (ecosystem-artifact, via its
			// rest segment) — a single segment fits neither.
			name: "no shape fits a one-segment path",
			ref:  "prn://acme/onlyone",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ref := resourceref.MustParse(tc.ref)
			err := p.Validate(ref)
			if !errors.Is(err, ErrNoMatch) {
				t.Fatalf("Validate(%q): err = %v, want wrapping ErrNoMatch", tc.ref, err)
			}
			if _, err := p.Split(ref); !errors.Is(err, ErrNoMatch) {
				t.Fatalf("Split(%q): err = %v, want wrapping ErrNoMatch", tc.ref, err)
			}
		})
	}
}

func TestValidateZeroValue(t *testing.T) {
	p := referenceProfile(t)
	var zero resourceref.ResourceRef
	if err := p.Validate(zero); !errors.Is(err, ErrNoMatch) {
		t.Fatalf("Validate(zero ResourceRef): err = %v, want wrapping ErrNoMatch", err)
	}
}

func TestVersionRequiredPolicy(t *testing.T) {
	// A profile identical to the fixture but with version.required=true,
	// to confirm the Required policy is actually enforced when set (the
	// fixture itself leaves it optional so the "no version" golden case
	// can exist).
	data := []byte(`{
		"scheme": "prn",
		"uriSubset": {"allowQuery": false, "allowFragment": false, "allowUserinfo": false, "allowPort": false},
		"authorities": [{"name": "acme"}],
		"shapes": [{"name": "s", "segments": [{"role": "resourceType"}, {"role": "id"}]}],
		"version": {"separator": "@", "required": true, "versionPattern": "^[0-9]+$"}
	}`)
	p, err := Load(data)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if err := p.Validate(resourceref.MustParse("prn://acme/invoice/nb_1@1")); err != nil {
		t.Fatalf("Validate(with version): unexpected error: %v", err)
	}
	if err := p.Validate(resourceref.MustParse("prn://acme/invoice/nb_1")); !errors.Is(err, ErrNoMatch) {
		t.Fatalf("Validate(without version, required=true): err = %v, want wrapping ErrNoMatch", err)
	}
}

func TestURISubsetComponentsRejected(t *testing.T) {
	p := referenceProfile(t)
	valid := "prn://acme/billing/org_1/invoice/nb_1"

	cases := []struct {
		name string
		ref  string
	}{
		{name: "fragment", ref: valid + "#frag"},
		{name: "userinfo", ref: "prn://user:pass@acme/billing/org_1/invoice/nb_1"}, // #nosec G101 -- test URI literal, not a credential
		{name: "port", ref: "prn://acme:8080/billing/org_1/invoice/nb_1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := p.Validate(resourceref.MustParse(tc.ref)); !errors.Is(err, ErrNoMatch) {
				t.Fatalf("Validate(%q): err = %v, want wrapping ErrNoMatch", tc.ref, err)
			}
		})
	}
}

func TestOpaqueReferenceHasNoAuthority(t *testing.T) {
	p := referenceProfile(t)
	// Same scheme as the profile, but opaque (no "//"): matches Scheme,
	// then must fail at the authority step, not crash on a nil host.
	if err := p.Validate(resourceref.MustParse("prn:opaque-thing")); !errors.Is(err, ErrNoMatch) {
		t.Fatalf("Validate(opaque prn:): err = %v, want wrapping ErrNoMatch", err)
	}
}

func TestEmptyPathAfterAuthority(t *testing.T) {
	p := referenceProfile(t)
	if err := p.Validate(resourceref.MustParse("prn://acme")); !errors.Is(err, ErrNoMatch) {
		t.Fatalf("Validate(authority-only, no path): err = %v, want wrapping ErrNoMatch", err)
	}
}

func TestAuthorityFailingSegmentPattern(t *testing.T) {
	// Load doesn't cross-validate authority names against segmentPattern
	// (they're independent concerns at authoring time), so an authority
	// that itself violates the pattern is a legitimate — if unusual —
	// profile, and must still be caught at match time rather than let a
	// non-conforming reference through.
	data := []byte(`{
		"scheme": "prn",
		"uriSubset": {
			"allowQuery": false, "allowFragment": false, "allowUserinfo": false, "allowPort": false,
			"segmentPattern": "^[a-z][a-z0-9-]*$"
		},
		"authorities": [{"name": "123bad"}],
		"shapes": [{"name": "s", "segments": [{"role": "resourceType"}, {"role": "id"}]}],
		"version": {}
	}`)
	p, err := Load(data)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := p.Validate(resourceref.MustParse("prn://123bad/invoice/nb_1")); !errors.Is(err, ErrNoMatch) {
		t.Fatalf("Validate: err = %v, want wrapping ErrNoMatch", err)
	}
}

func TestVersionSuffixEdgeCases(t *testing.T) {
	data := []byte(`{
		"scheme": "prn",
		"uriSubset": {"allowQuery": false, "allowFragment": false, "allowUserinfo": false, "allowPort": false},
		"authorities": [{"name": "a"}],
		"shapes": [{"name": "s", "segments": [{"role": "resourceType"}, {"role": "id"}]}],
		"version": {"separator": "@", "versionPattern": "^[0-9]+$", "aliasPattern": "^[a-z][a-z0-9-]*$"}
	}`)
	p, err := Load(data)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	cases := []struct {
		name string
		ref  string
	}{
		{name: "empty segment before separator", ref: "prn://a/invoice/@prod"},
		{name: "empty suffix after separator", ref: "prn://a/invoice/nb_1@"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := p.Validate(resourceref.MustParse(tc.ref)); !errors.Is(err, ErrNoMatch) {
				t.Fatalf("Validate(%q): err = %v, want wrapping ErrNoMatch", tc.ref, err)
			}
		})
	}
}

func TestMatchTimeGuardsAgainstHandBuiltProfile(t *testing.T) {
	// Load rejects an invalid regex before a Profile ever exists, but a
	// caller can bypass Load entirely by constructing a Profile{} literal
	// directly. match() must still fail cleanly with ErrInvalidProfile —
	// not panic — rather than assume every Profile it sees came from Load.
	segPattern := Profile{
		Scheme:      "prn",
		URISubset:   URISubset{SegmentPattern: "(unclosed"},
		Authorities: []Authority{{Name: "a"}},
		Shapes:      []Shape{{Name: "s", Segments: []SegmentSpec{{Role: RoleResourceType}, {Role: RoleID}}}},
	}
	if err := segPattern.Validate(resourceref.MustParse("prn://a/invoice/nb_1")); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("Validate with bad segmentPattern: err = %v, want wrapping ErrInvalidProfile", err)
	}

	versionPattern := Profile{
		Scheme:      "prn",
		Authorities: []Authority{{Name: "a"}},
		Shapes:      []Shape{{Name: "s", Segments: []SegmentSpec{{Role: RoleResourceType}, {Role: RoleID}}}},
		Version:     VersionGrammar{Separator: "@", VersionPattern: "(unclosed"},
	}
	if err := versionPattern.Validate(resourceref.MustParse("prn://a/invoice/nb_1@1")); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("Validate with bad versionPattern: err = %v, want wrapping ErrInvalidProfile", err)
	}
}

func TestShapeOrderFirstMatchWins(t *testing.T) {
	// Two shapes that both fit a 2-segment path (each valid on its own —
	// both include a resourceType segment, as Load requires — but
	// assigning roles in the opposite order), so only declaration order
	// distinguishes which one Split actually applies.
	data := []byte(`{
		"scheme": "prn",
		"uriSubset": {"allowQuery": false, "allowFragment": false, "allowUserinfo": false, "allowPort": false},
		"authorities": [{"name": "a"}],
		"shapes": [
			{"name": "first", "segments": [{"role": "resourceType"}, {"role": "id"}]},
			{"name": "second", "segments": [{"role": "id"}, {"role": "resourceType"}]}
		],
		"version": {}
	}`)
	p, err := Load(data)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, err := p.Split(resourceref.MustParse("prn://a/x/y"))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	want := Parts{Authority: "a", ResourceType: "x", ID: "y"}
	if got != want {
		t.Fatalf("Split = %+v, want %+v (first-declared shape should win)", got, want)
	}
}
