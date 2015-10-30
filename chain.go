package xhandler

import (
	"net/http"

	"golang.org/x/net/context"
)

// Chain is an helper to chain middleware handlers together for an easier
// management.
type Chain []func(next HandlerC) HandlerC

// UseC appends a context-aware handler to the middeware chain.
func (c *Chain) UseC(f func(next HandlerC) HandlerC) {
	*c = append(*c, f)
}

// Use appends a standard http.Handler to the middleware chain without
// lossing track of the context when inserted between two context aware handlers.
func (c *Chain) Use(f func(next http.Handler) http.Handler) {
	xf := func(next HandlerC) HandlerC {
		var rw http.ResponseWriter
		var rr *http.Request
		n := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw = w
			rr = r
		})
		s := f(n).ServeHTTP
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			s(w, r)
			next.ServeHTTPC(ctx, rw, rr)
		})
	}
	*c = append(*c, xf)
}

// Handler wraps the provided final handler with all the middleware appended to
// the chain and return a new standard http.Handler instance.
// The context.Background() context is injected automatically.
func (c Chain) Handler(xh HandlerC) http.Handler {
	ctx := context.Background()
	return c.HandlerCtx(ctx, xh)
}

// HandlerCtx wraps the provided final handler with all the middleware appended to
// the chain and return a new standard http.Handler instance.
func (c Chain) HandlerCtx(ctx context.Context, xh HandlerC) http.Handler {
	for i := len(c) - 1; i >= 0; i-- {
		xh = c[i](xh)
	}
	return New(ctx, xh)
}
