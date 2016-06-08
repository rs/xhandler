// Package xhandler provides a bridge between http.Handler and net/context.
//
// xhandler enforces net/context in your handlers without sacrificing
// compatibility with existing http.Handlers nor imposing a specific router.
//
// Thanks to net/context deadline management, xhandler is able to enforce
// a per request deadline and will cancel the context in when the client close
// the connection unexpectedly.
//
// You may create net/context aware middlewares pretty much the same way as
// you would do with http.Handler.
package xhandler // import "github.com/rs/xhandler"

import (
	"context"
	"net/http"
)

// New creates a conventional http.Handler injecting the provided root
// context to sub handlers. This handler is used as a bridge between conventional
// http.Handler and context aware handlers.
func New(ctx context.Context, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ctx = context.WithContext(ctx)
		r = r.WithContext(ctx) // TODO: doesnt smell good.. might loose orig ctx?!
		h.ServeHTTP(w, r)
	})
}
