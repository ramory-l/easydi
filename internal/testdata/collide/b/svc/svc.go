package svc

import asvc "github.com/ramory-l/easydi/internal/testdata/collide/a/svc"

type B struct{ A *asvc.A }

// di:provide name=B
// di:expose
func NewB(a *asvc.A) *B { return &B{A: a} }
