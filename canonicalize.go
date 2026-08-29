package resourceref

import (
	"net/url"
	"strings"
)

// canonicalize renders a parsed *url.URL in canonical form, applying only
// the conservative transformations documented on Parse. It never touches
// components it cannot safely reinterpret.
func canonicalize(u *url.URL) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(u.Scheme))
	b.WriteByte(':')

	switch {
	case u.Opaque != "":
		// Rootless/opaque reference (e.g. "arn:...", "urn:...", "tel:...",
		// "mailto:..."): Go exposes no authority or path structure to
		// reinterpret, so only percent-encoding is normalized.
		b.WriteString(normalizePercentEncoding(u.Opaque))
	default:
		if !u.OmitHost {
			// A "//" authority was present in the original text (host may
			// still be empty, e.g. "file:///etc/passwd").
			//
			// u.Host is written back as a whole rather than decomposed via
			// Hostname()/Port(): those methods apply a best-effort
			// rightmost-colon split that is not guaranteed to round-trip
			// for a malformed (non-bracketed, colon-containing) host, which
			// breaks the idempotence invariant this package guarantees.
			// Lowercasing the raw host is always safe — host is
			// case-insensitive per RFC 3986 §3.2.2 — regardless of whether
			// it parses cleanly into a host/port pair.
			b.WriteString("//")
			if u.User != nil {
				b.WriteString(normalizePercentEncoding(u.User.String()))
				b.WriteByte('@')
			}
			b.WriteString(strings.ToLower(u.Host))
		}
		path := removeDotSegments(normalizePercentEncoding(u.EscapedPath()))
		if u.OmitHost && strings.HasPrefix(path, "//") {
			// RFC 3986 §3.3: a no-authority URI's path must not begin with
			// "//" — "scheme:" followed by a path starting "//" is
			// indistinguishable from "scheme://" followed by an authority.
			// Dot-segment removal can produce exactly that shape (e.g. the
			// path "/.//A" removes its "/./" segment down to "//A"), so
			// re-escape with the standard "/." prefix to keep the
			// serialized form unambiguous and, in turn, idempotent: a
			// second Parse must not reinterpret this path as an authority.
			path = "/." + path
		}
		b.WriteString(path)
	}

	if u.ForceQuery || u.RawQuery != "" {
		b.WriteByte('?')
		b.WriteString(normalizePercentEncoding(u.RawQuery))
	}
	if u.Fragment != "" || u.RawFragment != "" {
		b.WriteByte('#')
		b.WriteString(normalizePercentEncoding(u.EscapedFragment()))
	}
	return b.String()
}

// normalizePercentEncoding applies RFC 3986 §6.2.2.2: any percent-encoded
// octet that corresponds to an unreserved character (ALPHA / DIGIT / "-" /
// "." / "_" / "~") is decoded to that literal character; any other
// percent-encoding is left encoded but with its hex digits uppercased.
// Bytes outside of "%XX" sequences pass through unchanged.
func normalizePercentEncoding(s string) string {
	if !strings.ContainsRune(s, '%') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '%' {
			b.WriteByte(c)
			continue
		}
		if i+2 < len(s) && isHexDigit(s[i+1]) && isHexDigit(s[i+2]) {
			decoded := hexVal(s[i+1])<<4 | hexVal(s[i+2])
			if isUnreserved(decoded) {
				b.WriteByte(decoded)
			} else {
				b.WriteByte('%')
				b.WriteByte(upperHex(s[i+1]))
				b.WriteByte(upperHex(s[i+2]))
			}
			i += 2
			continue
		}
		// A '%' not followed by two hex digits is not a valid pct-encoded
		// triplet (RFC 3986 permits this in net/url's unvalidated Opaque
		// component). Emitting it as a literal '%' would let the next
		// normalization pass combine it with whatever follows into a
		// *different* triplet than this pass saw — an alignment shift
		// that breaks idempotence. Escaping it to its own encoding is a
		// stable fixed point: '%' (0x25) is not unreserved, so "%25" is
		// never decoded back to a literal '%' by this function.
		b.WriteString("%25")
	}
	return b.String()
}

func isUnreserved(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') ||
		b == '-' || b == '.' || b == '_' || b == '~'
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func hexVal(b byte) byte {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	default: // 'A'-'F'
		return b - 'A' + 10
	}
}

func upperHex(b byte) byte {
	if b >= 'a' && b <= 'f' {
		return b - 'a' + 'A'
	}
	return b
}

// removeDotSegments implements the RFC 3986 §5.2.4 algorithm for removing
// "." and ".." segments from a URI path, without the caller needing a full
// base-URI resolution context.
func removeDotSegments(path string) string {
	in, out := path, ""
	for in != "" {
		switch {
		case strings.HasPrefix(in, "../"):
			in = in[3:]
		case strings.HasPrefix(in, "./"):
			in = in[2:]
		case strings.HasPrefix(in, "/./"):
			in = "/" + in[3:]
		case in == "/.":
			in = "/"
		case strings.HasPrefix(in, "/../"):
			in = "/" + in[4:]
			out = trimLastSegment(out)
		case in == "/..":
			in = "/"
			out = trimLastSegment(out)
		case in == "." || in == "..":
			in = ""
		default:
			start := 0
			if in[0] == '/' {
				start = 1
			}
			if j := strings.IndexByte(in[start:], '/'); j == -1 {
				out += in
				in = ""
			} else {
				out += in[:start+j]
				in = in[start+j:]
			}
		}
	}
	return out
}

// trimLastSegment removes the last "/segment" (or bare content with no
// leading slash) from an accumulated output path, per RFC 3986 §5.2.4 step
// 2D's "remove the last segment" instruction.
func trimLastSegment(s string) string {
	if idx := strings.LastIndexByte(s, '/'); idx != -1 {
		return s[:idx]
	}
	return ""
}
