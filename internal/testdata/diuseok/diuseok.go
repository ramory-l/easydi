// Package diuseok is a testdata fixture for the di:use parameter selector
// (positive case — no conflict providers).
package diuseok

// Lookup is a minimal interface used in this fixture.
type Lookup interface{ ID() string }

type lookup struct{}

func (lookup) ID() string { return "lookup" }

// W is the product of NewW.
type W struct{}

// di:provide name=UserService
func NewUserService() Lookup {
	return lookup{}
}

// di:provide name=W
func NewW(
	// di:use UserService
	u Lookup,
) *W { return &W{} }
