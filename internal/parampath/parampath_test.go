package parampath

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	p, err := Parse("Infra.DB.GetPostgresDB().GoquDB")
	if err != nil {
		t.Fatal(err)
	}
	want := []Seg{
		{Name: "Infra"},
		{Name: "DB"},
		{Name: "GetPostgresDB", Call: true},
		{Name: "GoquDB"},
	}
	if !reflect.DeepEqual(p.Segs, want) {
		t.Fatalf("got %+v want %+v", p.Segs, want)
	}

	p2, err := Parse("time.Now")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p2.Segs, []Seg{{Name: "time"}, {Name: "Now"}}) {
		t.Fatalf("got %+v", p2.Segs)
	}

	// Single segment is the whole-root case (e.g. di:param Config).
	p3, err := Parse("Config")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p3.Segs, []Seg{{Name: "Config"}}) {
		t.Fatalf("got %+v", p3.Segs)
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{
		"", "Config.", ".Auth", "Config..Auth",
		"Config.Make(x)", "Config.Arr[0]", "a + b", "Config.Auth ",
		"Get(1)", "1bad.X",
	} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected error", in)
		}
	}
}
