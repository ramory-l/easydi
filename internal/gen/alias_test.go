package gen

import (
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

func TestSegmentsToIdent(t *testing.T) {
	cases := map[string][]string{
		"twitchHttp":   {"twitch", "http"},
		"handlersAuth": {"handlers", "auth"},
		"goqu":         {"goqu"},
		"pV9":          {"v9"}, // leading digit -> prefixed
	}
	for want, segs := range cases {
		if got := segmentsToIdent(segs); got != want {
			t.Fatalf("segmentsToIdent(%v) = %q, want %q", segs, got, want)
		}
	}
}
