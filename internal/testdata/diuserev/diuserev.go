// Package diuserev is a testdata fixture for the di:use/di:param conflict
// with the directive lines stacked in the reverse order (di:use above
// di:param), exercising order-independence of the mutual-exclusion check.
package diuserev

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

// Bad is the product of NewBad.
type Bad struct{}

// di:provide name=Bad
func NewBad(
	// di:use UserService
	// di:param Config.X
	u Lookup,
) *Bad {
	return &Bad{}
}
