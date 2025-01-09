package model

type key int

const (
	// CtxContext context key for custom context object
	CtxContext key = iota
	CtxUser
)
