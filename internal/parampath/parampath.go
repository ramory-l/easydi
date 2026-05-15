// Package parampath parses and validates an easydi di:param path.
//
// Grammar (bounded — not a DSL): ident ( "." ident [ "()" ] )*
// i.e. dotted identifiers where any non-head segment may be a zero-arg call.
package parampath

import (
	"fmt"
	"strings"
)

// Seg is one dotted segment of a di:param path.
type Seg struct {
	Name string
	Call bool // segment is a zero-arg method call, e.g. GetPostgresDB()
}

// Path is a parsed, validated di:param path: an ordered list of segments.
type Path struct {
	Segs []Seg
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		ok := r == '_' ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(i > 0 && r >= '0' && r <= '9')
		if !ok {
			return false
		}
	}
	return true
}

// Parse parses and validates a di:param path string against the bounded
// grammar. A single-segment path is the whole-root case; classifying a head
// as a root versus a package-qualified literal is left to the resolver.
func Parse(s string) (Path, error) {
	if s == "" || s != strings.TrimSpace(s) {
		return Path{}, fmt.Errorf("parampath: empty or padded path %q", s)
	}
	parts := strings.Split(s, ".")
	segs := make([]Seg, 0, len(parts))
	for i, raw := range parts {
		seg := Seg{}
		name := raw
		if strings.HasSuffix(raw, "()") {
			if i == 0 {
				return Path{}, fmt.Errorf("parampath: head %q cannot be a call", raw)
			}
			seg.Call = true
			name = strings.TrimSuffix(raw, "()")
		}
		if strings.ContainsAny(name, "()[]{}+-*/ ,") {
			return Path{}, fmt.Errorf("parampath: illegal segment %q", raw)
		}
		if !isIdent(name) {
			return Path{}, fmt.Errorf("parampath: %q is not an identifier", raw)
		}
		seg.Name = name
		segs = append(segs, seg)
	}
	if len(segs) == 0 {
		return Path{}, fmt.Errorf("parampath: empty path %q", s)
	}
	// A single segment is the whole-root case (e.g. `di:param Config`);
	// classification root-vs-literal happens in the resolver.
	return Path{Segs: segs}, nil
}

// String renders the path back to Go source (used in code generation,
// prefixed by the resolved root variable).
func (p Path) String() string {
	var b strings.Builder
	for i, s := range p.Segs {
		if i > 0 {
			b.WriteByte('.')
		}
		b.WriteString(s.Name)
		if s.Call {
			b.WriteString("()")
		}
	}
	return b.String()
}
