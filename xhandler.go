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
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// Handler is a net/context aware http.Handler
type Handler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc type is an adapter to allow the use of ordinary functions
// as a xhandler.Handler. If f is a function with the appropriate signature,
// xhandler.HandlerFunc(f) is a xhandler.Handler object that calls f.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTP calls f(ctx, w, r).
func (f HandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

// CtxHandler creates a conventional http.Handler injecting the provided root
// context to sub heandlers. This handler is used as a bridge between conventional
// http.Handler and context aware handlers.
func CtxHandler(ctx context.Context, h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(ctx, w, r)
	})
}

// CloseHandler returns a Handler cancelling the context when the client
// connection close unexpectedly.
func CloseHandler(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Cancel the context if the client closes the connection
		if wcn, ok := w.(http.CloseNotifier); ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			defer cancel()

			notify := wcn.CloseNotify()
			go func() {
				<-notify
				cancel()
			}()
		}

		h.ServeHTTP(ctx, w, r)
	})
}

// TimeoutHandler returns a Handler which adds a timeout to the context.
//
// Child handlers have the responsability to obey the context deadline and to return
// an appropriate error (or not) response in case of timeout.
func TimeoutHandler(h Handler, timeout time.Duration) Handler {
	return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx, _ = context.WithTimeout(ctx, timeout)
		h.ServeHTTP(ctx, w, r)
	})
}
