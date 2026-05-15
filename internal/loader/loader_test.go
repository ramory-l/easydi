package loader

import "testing"

func TestLoad(t *testing.T) {
	pkgs, err := Load("../testdata/simple")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("got %d packages", len(pkgs))
	}
	p := pkgs[0]
	if p.Types == nil || p.TypesInfo == nil || len(p.Syntax) == 0 {
		t.Fatalf("package missing type/syntax info: %s", p.Name)
	}
	if p.Types.Scope().Lookup("NewService") == nil {
		t.Fatalf("NewService not found in scope")
	}
}
