// Package diusenone is a testdata fixture for di:use referencing an unknown
// provider node name — should produce a "no provider node named" error.
package diusenone

// Lookup is the interface used in this fixture.
type Lookup interface{ ID() string }

type lookup struct{}

func (lookup) ID() string { return "lookup" }

// di:provide name=UserService
func NewUserService() Lookup { return lookup{} }

// Consumer is the product of NewConsumer.
type Consumer struct{}

// di:provide name=Consumer
func NewConsumer(
	// di:use Nope
	u Lookup,
) *Consumer { return &Consumer{} }
