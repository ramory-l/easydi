package scanner

import (
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
