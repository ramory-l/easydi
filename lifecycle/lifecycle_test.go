package lifecycle_test

import (
	"context"
	"testing"

	"github.com/ramory-l/easydi/lifecycle"
)

type both struct {
	started, closed bool
}

func (b *both) Start(context.Context) error { b.started = true; return nil }
func (b *both) Close(context.Context) error { b.closed = true; return nil }

func TestInterfacesAreSatisfiable(t *testing.T) {
	var (
		s lifecycle.Starter
		c lifecycle.Closer
	)
	x := &both{}
	s = x
	c = x
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := c.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !x.started || !x.closed {
		t.Fatalf("expected started+closed, got %+v", x)
	}
}
