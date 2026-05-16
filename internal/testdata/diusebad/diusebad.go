// Package diusebad is a testdata fixture for di:use where the named provider's
// produced type is not assignable to the parameter type — should produce a
// "not assignable to parameter" error.
//
// UserService produces Lookup (an interface). The parameter wants *Consumer
// (an unrelated concrete pointer type), so the assignability check must fail.
package diusebad

// Lookup is the interface produced by UserService.
type Lookup interface{ ID() string }

type lookup struct{}

func (lookup) ID() string { return "lookup" }

// di:provide name=UserService
func NewUserService() Lookup { return lookup{} }

// Consumer is the product of NewConsumer.
type Consumer struct{}

// di:provide name=Consumer
func NewConsumer(
	// di:use UserService
	u *Consumer,
) *Consumer { return u }
