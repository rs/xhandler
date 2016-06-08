package xhandler

import "net/http"

// Chain is an helper to chain middleware handlers together for an easier
// management.
type Chain []func(http.Handler) http.Handler

// Use appends a context-aware handler to the middleware chain.
//
// Caveat: the f function will be called on each request so you are better to put
// any initialization sequence outside of this function.
func (c *Chain) Use(f func(next http.Handler) http.Handler) {
	*c = append(*c, f)
}

// Handler wraps the provided final handler with all the middleware appended to
// the chain and return a new standard http.Handler instance.
// The context.Background() context is injected automatically.
func (c Chain) Handler(h http.Handler) http.Handler {
	for i := len(c) - 1; i >= 0; i-- {
		h = c[i](h)
	}
	return h
}

// HandlerF is an helper to provide a standard http handler function
// (http.HandlerFunc) to Handler().
func (c Chain) HandlerF(hf http.HandlerFunc) http.Handler {
	return c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hf(w, r)
	}))
}
