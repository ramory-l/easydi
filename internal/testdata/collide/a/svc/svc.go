package svc

type A struct{}

// di:provide name=A
func NewA() *A { return &A{} }
