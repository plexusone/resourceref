package resourceref

import "testing"

// FuzzParseIdempotent checks the core canonicalization invariant: the
// canonical form is a fixed point. If s parses successfully, reparsing its
// canonical form must succeed and produce the identical canonical form.
// This is what makes ResourceRef safe to use as a comparison/equality key
// regardless of how many times a reference has round-tripped through
// storage or another system.
func FuzzParseIdempotent(f *testing.F) {
	for _, s := range passthroughCorpus {
		f.Add(s)
	}
	seeds := []string{
		"https://example.com/a/b/../c?q=%68ello#frag",
		"HTTPS://Example.COM/%7Euser/./x/../y",
		"file:///etc/passwd",
		"file:/etc/passwd",
		"",
		"not a uri",
		"://bad",
		"a:",
		"a://",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		r1, err := Parse(s)
		if err != nil {
			return // invalid input is expected to occur under fuzzing
		}
		r2, err := Parse(r1.String())
		if err != nil {
			t.Fatalf("canonical form failed to reparse: %q: %v", r1.String(), err)
		}
		if r1.String() != r2.String() {
			t.Fatalf("canonicalization not idempotent: %q -> %q -> %q", s, r1, r2)
		}
	})
}
