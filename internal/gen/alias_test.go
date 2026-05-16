package gen

import (
	"go/token"
	"reflect"
	"testing"
)

func TestComputeAliases(t *testing.T) {
	// name[path] = the Go package identifier (p.Name()) for that import path.
	name := map[string]string{
		"example.com/m/internal/pkg/jwt":           "jwt",
		"example.com/m/internal/domains/user":      "user",
		"example.com/m/internal/handlers/user":     "user",
		"example.com/m/internal/repositories/user": "user",
		"example.com/m/internal/pkg/twitch/http":   "http",
		"example.com/m/internal/pkg/vklive/http":   "http",
		"example.com/m/internal/auth":              "auth",
		"example.com/m/internal/handlers/auth":     "auth",
		"context":                                  "context",
	}
	paths := make([]string, 0, len(name))
	for p := range name {
		paths = append(paths, p)
	}

	got := computeAliases(paths, name)

	want := map[string]string{
		"example.com/m/internal/pkg/jwt":           "jwt",     // unique -> bare
		"context":                                  "context", // unique -> bare
		"example.com/m/internal/domains/user":      "domainsUser",
		"example.com/m/internal/handlers/user":     "handlersUser",
		"example.com/m/internal/repositories/user": "repositoriesUser",
		"example.com/m/internal/pkg/twitch/http":   "twitchHttp",
		"example.com/m/internal/pkg/vklive/http":   "vkliveHttp",
		"example.com/m/internal/auth":              "internalAuth",
		"example.com/m/internal/handlers/auth":     "handlersAuth",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("computeAliases mismatch\n got: %#v\nwant: %#v", got, want)
	}

	// Determinism: identical input (any order) -> identical output.
	got2 := computeAliases([]string{
		"context",
		"example.com/m/internal/handlers/auth",
		"example.com/m/internal/auth",
		"example.com/m/internal/pkg/jwt",
		"example.com/m/internal/repositories/user",
		"example.com/m/internal/domains/user",
		"example.com/m/internal/handlers/user",
		"example.com/m/internal/pkg/vklive/http",
		"example.com/m/internal/pkg/twitch/http",
	}, name)
	if !reflect.DeepEqual(got2, want) {
		t.Fatalf("computeAliases not deterministic\n got: %#v\nwant: %#v", got2, want)
	}
}

func TestComputeAliasesKeywordAndCollision(t *testing.T) {
	name := map[string]string{
		"example.com/a/type":   "type",   // keyword as package ident
		"example.com/b/select": "select", // keyword
	}
	got := computeAliases([]string{"example.com/a/type", "example.com/b/select"}, name)
	// Singletons, but the bare name is a Go keyword -> must be aliased.
	if got["example.com/a/type"] == "type" || got["example.com/b/select"] == "select" {
		t.Fatalf("keyword package idents must be aliased: %#v", got)
	}
	for _, a := range got {
		if a == "type" || a == "select" {
			t.Fatalf("alias must not be a keyword: %#v", got)
		}
	}
}

// M2: converted to ordered slice so all failures are reported via t.Errorf.
func TestSegmentsToIdent(t *testing.T) {
	cases := []struct {
		segs []string
		want string
	}{
		{[]string{"twitch", "http"}, "twitchHttp"},
		{[]string{"handlers", "auth"}, "handlersAuth"},
		{[]string{"goqu"}, "goqu"},
		{[]string{"v9"}, "pV9"}, // leading digit -> prefixed
	}
	for _, tc := range cases {
		if got := segmentsToIdent(tc.segs); got != tc.want {
			t.Errorf("segmentsToIdent(%v) = %q, want %q", tc.segs, got, tc.want)
		}
	}
}

// M3: uniqueSuffixAlias numeric fallback — every suffix depth is pre-occupied.
func TestUniqueSuffixAliasNumericFallback(t *testing.T) {
	path := "example.com/auth"
	// Pre-occupy all suffix candidates: "auth" and "exampleComAuth" (full path ident).
	used := map[string]bool{
		"auth":           true,
		"exampleComAuth": true,
	}
	got := uniqueSuffixAlias(path, used)
	if got == "" {
		t.Fatal("uniqueSuffixAlias returned empty string")
	}
	if used[got] {
		t.Fatalf("uniqueSuffixAlias returned already-used alias %q", got)
	}
	if token.IsKeyword(got) {
		t.Fatalf("uniqueSuffixAlias returned keyword alias %q", got)
	}
}

// M4: cross-group collision fallback — minimal-suffix alias equals a bare alias
// already assigned to a different singleton group, all final aliases must be unique.
func TestComputeAliasesCrossGroupCollision(t *testing.T) {
	// "handlersUser" is the singleton bare alias for "example.com/handlersUser".
	// The collision group {handlers/user, domains/user} will want "handlersUser"
	// for the handlers/user member at k=1 -> k=2, but "handlersUser" is already used.
	name := map[string]string{
		"example.com/handlersUser":             "handlersUser", // singleton, bare alias = "handlersUser"
		"example.com/m/internal/handlers/user": "user",
		"example.com/m/internal/domains/user":  "user",
	}
	paths := []string{
		"example.com/handlersUser",
		"example.com/m/internal/handlers/user",
		"example.com/m/internal/domains/user",
	}
	got := computeAliases(paths, name)

	seen := map[string]bool{}
	for p, a := range got {
		if a == "" {
			t.Errorf("path %q got empty alias", p)
		}
		if seen[a] {
			t.Errorf("duplicate alias %q", a)
		}
		seen[a] = true
		if token.IsKeyword(a) {
			t.Errorf("alias %q for %q is a keyword", a, p)
		}
	}
}

// Critical regression: paths differing only in all-non-identifier segments must
// not cause computeAliases to loop forever; it must return with distinct aliases.
func TestComputeAliasesIndistinguishableSegments(t *testing.T) {
	// sanitize("---") == "" and sanitize("!!!") == "", so at every suffix depth
	// both paths yield identical segmentsToIdent output.  The loop must terminate.
	name := map[string]string{
		"company/---/auth": "auth",
		"company/!!!/auth": "auth",
	}
	paths := []string{"company/---/auth", "company/!!!/auth"}
	got := computeAliases(paths, name)

	if len(got) != 2 {
		t.Fatalf("expected 2 aliases, got %d: %#v", len(got), got)
	}
	a1 := got["company/---/auth"]
	a2 := got["company/!!!/auth"]
	if a1 == "" || a2 == "" {
		t.Fatalf("got empty alias: %#v", got)
	}
	if a1 == a2 {
		t.Fatalf("aliases must be distinct, both got %q", a1)
	}
	if token.IsKeyword(a1) || token.IsKeyword(a2) {
		t.Fatalf("aliases must not be keywords: %q, %q", a1, a2)
	}
}
