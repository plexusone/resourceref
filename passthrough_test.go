package resourceref

import "testing"

// passthroughCorpus holds owner-canonical external references — an AWS
// ARN, a SPIFFE ID, an HTTPS URL, and URNs — that are already in their
// canonical form. The fidelity invariant these enforce is the one the PRD
// leads with: a resource reference is a URI, and this package must never
// rewrite an owner's own canonical reference. Parse(s).String() must
// reproduce s exactly.
var passthroughCorpus = []string{
	// AWS ARN
	"arn:aws:s3:::customer-data",
	"arn:aws:iam::123456789012:role/payment-api",
	"arn:aws:lambda:us-west-2:123456789012:function:my-function",

	// SPIFFE ID
	"spiffe://example.org/ns/prod/sa/api",
	"spiffe://prod.example.com/ns/api/sa/default",

	// HTTPS URL (GitHub, as used elsewhere in the ecosystem for lineage)
	"https://github.com/plexusone/dashforge",
	"https://github.com/plexusone/dashforge/blob/main/README.md",

	// URN (RFC 8141)
	"urn:isbn:0451450523",
	"urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	"urn:ietf:params:scim:schemas:core:2.0:User",

	// Other opaque schemes seen in the wild
	"mailto:jane@example.com",
	"tel:+1-816-555-1212",

	// PlexusOne's own owned scheme (defined in plexusone-standards; this
	// package validates only URI syntax, not scheme-specific structure)
	"prn://plexusone/dashforge/org_7f3/notebook/nb_churn@prod",
	"prn://productbuildershq/roadmap/RMI-142",
}

func TestPassthroughFidelity(t *testing.T) {
	for _, s := range passthroughCorpus {
		t.Run(s, func(t *testing.T) {
			r, err := Parse(s)
			if err != nil {
				t.Fatalf("Parse(%q): unexpected error: %v", s, err)
			}
			if got := r.String(); got != s {
				t.Fatalf("Parse(%q).String() = %q, want unchanged %q", s, got, s)
			}
		})
	}
}

func TestPassthroughSchemesDistinct(t *testing.T) {
	want := map[string]string{
		"arn:aws:s3:::customer-data":                               "arn",
		"spiffe://example.org/ns/prod/sa/api":                      "spiffe",
		"https://github.com/plexusone/dashforge":                   "https",
		"urn:isbn:0451450523":                                      "urn",
		"prn://plexusone/dashforge/org_7f3/notebook/nb_churn@prod": "prn",
	}
	for s, wantScheme := range want {
		if got := MustParse(s).Scheme(); got != wantScheme {
			t.Errorf("Scheme(%q) = %q, want %q", s, got, wantScheme)
		}
	}
}
