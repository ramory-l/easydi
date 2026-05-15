package cycle

// di:provide
func NewA(b B) A { return A{} }

// di:provide
func NewB(a A) B { return B{} }

type A struct{}
type B struct{}
