// Package diuseambig is a testdata fixture for di:use parameter selection
// where two providers both satisfy an interface (ambiguous by type) and
// di:use selects one explicitly.
package diuseambig

// Lookup is a minimal interface satisfied by both providers.
type Lookup interface{ ID() string }

type userRepo struct{}

func (userRepo) ID() string { return "repo" }

type userService struct{}

func (userService) ID() string { return "service" }

// di:provide name=UserRepository
func NewUserRepository() Lookup { return userRepo{} }

// di:provide name=UserService
func NewUserService() Lookup { return userService{} }

// Consumer is the product of NewConsumer.
type Consumer struct{}

// di:provide name=Consumer
func NewConsumer(
	// di:use UserService
	u Lookup,
) *Consumer { return &Consumer{} }
