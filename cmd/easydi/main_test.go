package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The generated container imports the fixture package
// github.com/ramory-l/easydi/internal/testdata/simple. Go's internal-package
// rule only permits that import from code inside the easydi module, so the
// compile check generates into an in-module package
// (internal/testdata/built) and builds it from the module root. Generating
// into a separate temp module would violate the internal rule and never
// compile.
func TestGenerateAndCompile(t *testing.T) {
	root := repoRoot(t)
	builtDir := filepath.Join(root, "internal", "testdata", "built")
	// Do NOT pre-create the nested dir: run() must create it via MkdirAll.
	out := filepath.Join(builtDir, "nested", "easydi_gen.go")
	t.Cleanup(func() { _ = os.RemoveAll(builtDir) })

	if err := run([]string{"gen", "-o", out, "-pkg", "nested", "../../internal/testdata/simple"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("output not written (dir auto-create failed?): %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty output")
	}

	// internal/testdata/built/nested and internal/testdata/simple share the
	// easydi module root, so the internal import is allowed. An explicit
	// package path builds even though it is under a testdata/ directory.
	cmd := exec.Command("go", "build", "./internal/testdata/built/nested/")
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated code did not compile: %v\n%s", err, b)
	}
}

func TestDefaultOutputName(t *testing.T) {
	if defaultOut != "easydi_gen.go" {
		t.Fatalf("default output name = %q, want easydi_gen.go", defaultOut)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd() // .../easydi/cmd/easydi
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}
