package handler

type key int

const (
	// CtxKey context key for request key field
	CtxKey key = iota
	// CtxQuotaValue context key for quota value field
	CtxQuotaValue key = iota
)
