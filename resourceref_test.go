package resourceref

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseErrors(t *testing.T) {
	cases := []string{
		"",
		"/relative/path",
		"no-scheme-here",
		"://missing-scheme-name",
	}
	for _, s := range cases {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q): expected error, got nil", s)
		} else if !errors.Is(err, ErrNoScheme) && s != "://missing-scheme-name" {
			// "://missing-scheme-name" fails net/url's own scheme syntax
			// check rather than our explicit empty-scheme check; either
			// failure mode is acceptable, we just require an error.
			t.Logf("Parse(%q): error = %v (not ErrNoScheme, but still an error, ok)", s, err)
		}
	}
}

func TestParseZeroValue(t *testing.T) {
	var r ResourceRef
	if !r.IsZero() {
		t.Fatal("zero ResourceRef should report IsZero() == true")
	}
	if r.String() != "" {
		t.Fatalf("zero ResourceRef.String() = %q, want empty", r.String())
	}
	if r.Scheme() != "" {
		t.Fatalf("zero ResourceRef.Scheme() = %q, want empty", r.Scheme())
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustParse(invalid) did not panic")
		}
	}()
	MustParse("not a valid reference")
}

func TestScheme(t *testing.T) {
	r := MustParse("HTTPS://example.com/path")
	if got, want := r.Scheme(), "https"; got != want {
		t.Fatalf("Scheme() = %q, want %q", got, want)
	}
}

func TestEqual(t *testing.T) {
	a := MustParse("HTTPS://Example.com/a/b")
	b := MustParse("https://example.com/a/b")
	if !a.Equal(b) {
		t.Fatalf("expected %q to equal %q after canonicalization", a, b)
	}
	c := MustParse("https://example.com/a/c")
	if a.Equal(c) {
		t.Fatalf("did not expect %q to equal %q", a, c)
	}
}

func TestCanonicalization(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "scheme and host lowercased, path case preserved",
			in:   "HTTPS://GitHub.com/acme/BillingService",
			want: "https://github.com/acme/BillingService",
		},
		{
			name: "dot segment .. removed",
			in:   "https://example.com/a/b/../c",
			want: "https://example.com/a/c",
		},
		{
			name: "dot segment . removed",
			in:   "https://example.com/a/./b",
			want: "https://example.com/a/b",
		},
		{
			name: "leading .. climbs past root and is dropped",
			in:   "https://example.com/../a",
			want: "https://example.com/a",
		},
		{
			name: "unreserved percent-encoding decoded",
			in:   "https://example.com/%7Euser",
			want: "https://example.com/~user",
		},
		{
			name: "reserved percent-encoding kept but hex uppercased",
			in:   "https://example.com/a%2fb",
			want: "https://example.com/a%2Fb",
		},
		{
			name: "reserved percent-encoding never decoded to literal slash",
			in:   "https://example.com/%2Fa",
			want: "https://example.com/%2Fa",
		},
		{
			name: "query percent-encoding normalized",
			in:   "https://example.com/search?q=%68ello",
			want: "https://example.com/search?q=hello",
		},
		{
			name: "explicit port is preserved, no default-port stripping",
			in:   "https://example.com:443/",
			want: "https://example.com:443/",
		},
		{
			name: "userinfo preserved, only host lowercased",
			in:   "https://User:Pass@Example.com/", // #nosec G101 -- test URI literal, not a credential
			want: "https://User:Pass@example.com/",
		},
		{
			name: "opaque reference untouched beyond percent-encoding",
			in:   "arn:aws:iam::123456789012:role/payment-api",
			want: "arn:aws:iam::123456789012:role/payment-api",
		},
		{
			name: "opaque scheme lowercased, opaque body untouched",
			in:   "URN:isbn:0451450523",
			want: "urn:isbn:0451450523",
		},
		{
			name: "no authority marker, path-only form preserved",
			in:   "file:/etc/passwd",
			want: "file:/etc/passwd",
		},
		{
			name: "empty-host authority form preserved distinctly",
			in:   "file:///etc/passwd",
			want: "file:///etc/passwd",
		},
		{
			// Regression: reconstructing the authority via Hostname()/
			// Port() (Go's rightmost-colon heuristic) is not a stable
			// round trip for a malformed, non-bracketed colon-containing
			// host — found by FuzzParseIdempotent as "A://::" drifting to
			// "a://:" then "a://" across repeated Parse calls. Writing
			// u.Host back whole avoids the heuristic entirely.
			name: "malformed colon-containing host is stable, not decomposed",
			in:   "a://::",
			want: "a://::",
		},
		{
			// Regression: dot-segment removal on a no-authority path can
			// produce a path starting with "//", which serialized directly
			// after "scheme:" is indistinguishable from an authority
			// marker — found by FuzzParseIdempotent as "A:/.//A" drifting
			// through "a://A" (path, case preserved) to "a://a" (now
			// misread as authority "A", lowercased) across repeated Parse
			// calls. RFC 3986 §3.3 requires exactly this to never occur.
			name: "no-authority path is re-escaped to avoid an authority collision",
			in:   "A:/.//A",
			want: "a:/.//A",
		},
		{
			// Regression: a stray '%' not followed by two hex digits, if
			// left as a literal '%', can combine with subsequent bytes on
			// the next normalization pass into a different triplet than
			// this pass decoded — found by FuzzParseIdempotent as
			// "A:%%370" drifting through "a:%70" to "a:p". Escaping the
			// stray '%' to "%25" (never itself decoded, since '%' is not
			// unreserved) reaches a stable fixed point in one pass.
			name: "stray percent not forming a triplet is escaped, not left literal",
			in:   "A:%%370",
			want: "a:%2570",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q): unexpected error: %v", tc.in, err)
			}
			if got := r.String(); got != tc.want {
				t.Fatalf("Parse(%q).String() = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCanonicalizationIdempotent(t *testing.T) {
	inputs := []string{
		"HTTPS://GitHub.com/acme/BillingService",
		"https://example.com/a/b/../c",
		"https://example.com/%7Euser",
		"arn:aws:iam::123456789012:role/payment-api",
		"prn://acme/billing/org_7f3/invoice/inv_9c3@prod",
	}
	for _, in := range inputs {
		r1 := MustParse(in)
		r2 := MustParse(r1.String())
		if r1.String() != r2.String() {
			t.Fatalf("canonicalization not idempotent for %q: %q != %q", in, r1, r2)
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	type wrapper struct {
		Ref ResourceRef `json:"ref"`
	}

	w := wrapper{Ref: MustParse("prn://acme/billing/org_7f3/invoice/inv_9c3@prod")}
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	const want = `{"ref":"prn://acme/billing/org_7f3/invoice/inv_9c3@prod"}`
	if string(b) != want {
		t.Fatalf("json.Marshal(w) = %s, want %s", b, want)
	}

	var w2 wrapper
	if err := json.Unmarshal(b, &w2); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !w.Ref.Equal(w2.Ref) {
		t.Fatalf("round-tripped ref %q != original %q", w2.Ref, w.Ref)
	}
}

func TestJSONMarshalZeroValueErrors(t *testing.T) {
	var r ResourceRef
	if _, err := r.MarshalText(); err == nil {
		t.Fatal("expected MarshalText on zero ResourceRef to error")
	}
	type wrapper struct {
		Ref ResourceRef `json:"ref"`
	}
	if _, err := json.Marshal(wrapper{}); err == nil {
		t.Fatal("expected json.Marshal of zero ResourceRef field to error")
	}
}

func TestUnmarshalTextInvalid(t *testing.T) {
	var r ResourceRef
	if err := r.UnmarshalText([]byte("not a valid reference")); err == nil {
		t.Fatal("expected UnmarshalText to reject a schemeless string")
	}
}
