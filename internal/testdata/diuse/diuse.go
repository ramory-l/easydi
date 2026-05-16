// Package diuse is a testdata fixture for di:use parameter selection.
package diuse

// Lookup is a minimal interface used in this fixture.
type Lookup interface{ ID() string }

type lookup struct{}

func (lookup) ID() string { return "lookup" }

// Config is the di:root for the di:param conflict test.
//
// di:root
type Config struct {
	X string
}

// di:provide name=UserService
func NewUserService() Lookup {
	return lookup{}
}

// W is the product of NewW.
type W struct{}

// di:provide name=W
func NewW(
	// di:use UserService
	u Lookup,
) *W { return &W{} }

// Bad is the product of NewBad.
type Bad struct{}

// di:provide name=Bad
func NewBad(
	// di:param Config.X
	// di:use UserService
	u Lookup,
) *Bad { return &Bad{} }
