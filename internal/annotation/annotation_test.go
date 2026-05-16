package annotation

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		ok   bool
		want Directive
	}{
		{"di:provide", true, Directive{Kind: Provide}},
		{"di:provide name=allowlist", true, Directive{Kind: Provide, Name: "allowlist"}},
		{"di:root", true, Directive{Kind: Root}},
		{"di:param Config.Auth", true, Directive{Kind: Param, Path: "Config.Auth"}},
		{"di:param Infra.DB.GetPostgresDB().GoquDB", true, Directive{Kind: Param, Path: "Infra.DB.GetPostgresDB().GoquDB"}},
		{"di:expose", true, Directive{Kind: Expose}},
		{"not a directive", false, Directive{}},
		{"di:bogus", false, Directive{}},
	}
	for _, c := range cases {
		got, ok, err := Parse(c.in)
		if ok != c.ok {
			t.Fatalf("Parse(%q) ok=%v want %v (err=%v)", c.in, ok, c.ok, err)
		}
		if ok && got != c.want {
			t.Fatalf("Parse(%q)=%+v want %+v", c.in, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"di:provide name=", "di:param", "di:provide extra"} {
		if _, _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected error", in)
		}
	}
}

func TestParseDiUse(t *testing.T) {
	d, ok, err := Parse("di:use UserService")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if d.Kind != Use || d.Node != "UserService" {
		t.Fatalf("got %+v", d)
	}

	if _, _, err := Parse("di:use"); err == nil {
		t.Fatalf("di:use with no node must error")
	}
	if _, _, err := Parse("di:use A B"); err == nil {
		t.Fatalf("di:use with two args must error")
	}
}
