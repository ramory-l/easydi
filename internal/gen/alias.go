package gen

import (
	"go/token"
	"sort"
	"strings"
)

// computeAliases assigns a unique, deterministic import alias to every import
// path. pkgName maps each path to its Go package identifier (types.Package
// Name()). A path whose package identifier is unique across the whole set and
// is a valid non-keyword identifier keeps that bare name (emitted without an
// explicit alias). Colliding (or keyword) names get a deterministic alias
// built from the minimal unique suffix of path segments, lowerCamelCased.
func computeAliases(paths []string, pkgName map[string]string) map[string]string {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)

	groups := map[string][]string{}
	for _, p := range sorted {
		id := pkgName[p]
		groups[id] = append(groups[id], p)
	}

	groupNames := make([]string, 0, len(groups))
	for id := range groups {
		groupNames = append(groupNames, id)
	}
	sort.Strings(groupNames)

	alias := make(map[string]string, len(sorted))
	used := map[string]bool{}

	for _, id := range groupNames {
		members := groups[id] // globally sorted subset
		if len(members) == 1 && validBareIdent(id) && !used[id] {
			alias[members[0]] = id
			used[id] = true
			continue
		}
		// Collision group: find the minimum k such that all members produce
		// distinct identifiers at suffix depth k, then assign.
		segsPerMember := make([][]string, len(members))
		for i, p := range members {
			segsPerMember[i] = strings.Split(p, "/")
		}
		assigned := make([]string, len(members))
		for k := 1; ; k++ {
			cands := make([]string, len(members))
			ok := true
			seen := map[string]bool{}
			for i, segs := range segsPerMember {
				start := len(segs) - k
				if start < 0 {
					start = 0
				}
				c := segmentsToIdent(segs[start:])
				if seen[c] {
					ok = false
					break
				}
				seen[c] = true
				cands[i] = c
			}
			if ok {
				copy(assigned, cands)
				break
			}
		}
		// Now assign, resolving any collision with previously used aliases by
		// extending the suffix further for affected members.
		for i, p := range members {
			a := assigned[i]
			if used[a] || token.IsKeyword(a) {
				a = uniqueSuffixAlias(p, used)
			}
			alias[p] = a
			used[a] = true
		}
	}
	return alias
}

// validBareIdent reports whether s can be used as a bare import (a valid Go
// identifier that is not a keyword and not the blank identifier).
func validBareIdent(s string) bool {
	return s != "" && s != "_" && token.IsIdentifier(s) && !token.IsKeyword(s)
}

// uniqueSuffixAlias builds an identifier from the shortest trailing run of
// path segments that yields an identifier not already in used (and not a
// keyword). Falls back to the full path then a numeric suffix.
func uniqueSuffixAlias(path string, used map[string]bool) string {
	segs := strings.Split(path, "/")
	for k := 1; k <= len(segs); k++ {
		cand := segmentsToIdent(segs[len(segs)-k:])
		if !used[cand] && !token.IsKeyword(cand) {
			return cand
		}
	}
	base := segmentsToIdent(segs)
	for i := 2; ; i++ {
		cand := base + "_" + itoa(i)
		if !used[cand] && !token.IsKeyword(cand) {
			return cand
		}
	}
}

// segmentsToIdent lowerCamelCases path segments into one Go identifier:
// non-identifier runes dropped, first segment lowercased, subsequent segments
// Title-cased, a leading digit prefixed with "p".
func segmentsToIdent(segs []string) string {
	var b strings.Builder
	for i, s := range segs {
		clean := sanitize(s)
		if clean == "" {
			continue
		}
		if i == 0 || b.Len() == 0 {
			b.WriteString(strings.ToLower(clean[:1]) + clean[1:])
		} else {
			b.WriteString(strings.ToUpper(clean[:1]) + clean[1:])
		}
	}
	out := b.String()
	if out == "" {
		out = "pkg"
	}
	// Detect version-tag pattern: single letter followed by all digits (e.g.
	// "v9"). Treat as if the letter prefix is stripped and the digit leads, so
	// prefix with "p" and title-case the letter.
	if isVersionTag(out) {
		out = "p" + strings.ToUpper(out[:1]) + out[1:]
	} else if out[0] >= '0' && out[0] <= '9' {
		out = "p" + out
	}
	return out
}

// isVersionTag reports whether s looks like a Go module major-version suffix:
// exactly one ASCII letter followed by one or more ASCII digits (e.g. "v9",
// "v12").
func isVersionTag(s string) bool {
	if len(s) < 2 {
		return false
	}
	if !((s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z')) {
		return false
	}
	for _, c := range s[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '_' ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var d []byte
	for i > 0 {
		d = append([]byte{byte('0' + i%10)}, d...)
		i /= 10
	}
	return string(d)
}
