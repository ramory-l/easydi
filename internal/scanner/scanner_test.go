package scanner

import (
	"strings"
	"testing"

	"github.com/ramory-l/easydi/internal/loader"
)

func TestScan(t *testing.T) {
	pkgs, err := loader.Load("../testdata/simple")
	if err != nil {
		t.Fatal(err)
	}
	res, err := Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Roots) != 2 {
		t.Fatalf("roots=%d want 2", len(res.Roots))
	}
	if len(res.Providers) != 3 {
		t.Fatalf("providers=%d want 3", len(res.Providers))
	}

	byName := map[string]*Provider{}
	for _, p := range res.Providers {
		byName[p.Name] = p
	}

	svc := byName["NewService"]
	if svc == nil || !svc.Expose {
		t.Fatalf("NewService missing or not exposed: %+v", svc)
	}
	if !svc.ReturnsError {
		t.Fatalf("NewService should return error")
	}
	if len(svc.Params) != 2 {
		t.Fatalf("NewService params=%d", len(svc.Params))
	}
	// param 0 (r Repo) has no path; param 1 (sec) has di:param Config.Auth.Secret
	if svc.Params[0].Path != nil {
		t.Fatalf("param r should have no path")
	}
	if svc.Params[1].Path == nil || svc.Params[1].Path.String() != "Config.Auth.Secret" {
		t.Fatalf("param sec path=%+v", svc.Params[1].Path)
	}

	h := byName["NewHasher"]
	if h.Params[0].Path == nil || h.Params[0].Path.String() != "Infra.DB.DSN()" {
		t.Fatalf("NewHasher dsn path=%+v", h.Params[0].Path)
	}
}

// TestScanDiUse verifies that a // di:use <NodeName> comment directly above a
// parameter is scanned into Param.Use and that Param.Path remains nil.
func TestScanDiUse(t *testing.T) {
	pkgs, err := loader.Load("../testdata/diuseok")
	if err != nil {
		t.Fatal(err)
	}
	res, err := Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}

	byName := map[string]*Provider{}
	for _, p := range res.Providers {
		byName[p.Name] = p
	}

	w := byName["W"]
	if w == nil {
		t.Fatalf("provider W not found; got names: %v", func() []string {
			var ns []string
			for n := range byName {
				ns = append(ns, n)
			}
			return ns
		}())
	}
	if len(w.Params) != 1 {
		t.Fatalf("W.Params count=%d want 1", len(w.Params))
	}
	param := w.Params[0]
	if param.Use != "UserService" {
		t.Fatalf("W.Params[0].Use=%q want %q", param.Use, "UserService")
	}
	if param.Path != nil {
		t.Fatalf("W.Params[0].Path should be nil, got %+v", param.Path)
	}
}

// TestScanDiUseParamConflict verifies that a parameter with both // di:param
// and // di:use stacked directly above it causes Scan to return a
// mutual-exclusion error.
func TestScanDiUseParamConflict(t *testing.T) {
	pkgs, err := loader.Load("../testdata/diuse")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Scan(pkgs)
	if err == nil {
		t.Fatal("expected error for di:param + di:use conflict, got nil")
	}
	const want = "di:param and di:use are mutually exclusive"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not contain %q", err.Error(), want)
	}
}

func TestScanDiUseParamConflictReverseOrder(t *testing.T) {
	pkgs, err := loader.Load("../testdata/diuserev")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Scan(pkgs)
	if err == nil {
		t.Fatal("expected error for di:use + di:param conflict (reverse order), got nil")
	}
	const want = "di:param and di:use are mutually exclusive"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not contain %q", err.Error(), want)
	}
}
