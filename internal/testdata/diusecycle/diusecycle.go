// Package diusecycle is a testdata fixture for a dependency cycle introduced
// via di:use: NodeA uses NodeB and NodeB uses NodeA.
package diusecycle

// A is the product of NewA.
type A struct{}

// B is the product of NewB.
type B struct{}

// di:provide name=NodeA
func NewA(
	// di:use NodeB
	b *B,
) *A { return &A{} }

// di:provide name=NodeB
func NewB(
	// di:use NodeA
	a *A,
) *B { return &B{} }
