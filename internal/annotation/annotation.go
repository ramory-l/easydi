// Package annotation parses a single easydi `// di:` directive line.
package annotation

import (
	"fmt"
	"strings"
)

// Kind identifies which di directive a parsed line represents.
type Kind int

// The recognized di directive kinds.
const (
	Provide Kind = iota // di:provide [name=X]
	Root                // di:root
	Param               // di:param <path>
	Expose              // di:expose
	Use                 // di:use <NodeName>
)

// Directive is a parsed di comment line.
type Directive struct {
	Kind Kind
	Name string // for `di:provide name=X`
	Path string // for `di:param <path>`
	Node string // for `di:use <NodeName>`
}

// Parse parses the text of a single comment line (already stripped of the
// leading `//` and surrounding spaces). ok is false when the line is not a
// di directive at all; err is non-nil when it is a di directive but malformed.
func Parse(text string) (Directive, bool, error) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "di:") {
		return Directive{}, false, nil
	}
	body := strings.TrimPrefix(text, "di:")
	fields := strings.Fields(body)
	if len(fields) == 0 {
		return Directive{}, false, fmt.Errorf("empty di directive")
	}
	switch fields[0] {
	case "provide":
		d := Directive{Kind: Provide}
		switch len(fields) {
		case 1:
		case 2:
			if !strings.HasPrefix(fields[1], "name=") {
				return Directive{}, true, fmt.Errorf("di:provide: unknown arg %q", fields[1])
			}
			d.Name = strings.TrimPrefix(fields[1], "name=")
			if d.Name == "" {
				return Directive{}, true, fmt.Errorf("di:provide: empty name=")
			}
		default:
			return Directive{}, true, fmt.Errorf("di:provide: too many args")
		}
		return d, true, nil
	case "root":
		if len(fields) != 1 {
			return Directive{}, true, fmt.Errorf("di:root: takes no args")
		}
		return Directive{Kind: Root}, true, nil
	case "param":
		if len(fields) != 2 {
			return Directive{}, true, fmt.Errorf("di:param: expects exactly one path")
		}
		return Directive{Kind: Param, Path: fields[1]}, true, nil
	case "expose":
		if len(fields) != 1 {
			return Directive{}, true, fmt.Errorf("di:expose: takes no args")
		}
		return Directive{Kind: Expose}, true, nil
	case "use":
		if len(fields) != 2 {
			return Directive{}, true, fmt.Errorf("di:use: expects exactly one node name")
		}
		return Directive{Kind: Use, Node: fields[1]}, true, nil
	default:
		return Directive{}, false, nil
	}
}
