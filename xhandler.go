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

// HandlerC is a net/context aware http.Handler
type HandlerC interface {
	ServeHTTPC(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFuncC type is an adapter to allow the use of ordinary functions
// as a xhandler.Handler. If f is a function with the appropriate signature,
// xhandler.HandlerFunc(f) is a xhandler.Handler object that calls f.
type HandlerFuncC func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTPC calls f(ctx, w, r).
func (f HandlerFuncC) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

// New creates a conventional http.Handler injecting the provided root
// context to sub heandlers. This handler is used as a bridge between conventional
// http.Handler and context aware handlers.
func New(ctx context.Context, h HandlerC) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTPC(ctx, w, r)
	})
}

// CloseHandler returns a Handler cancelling the context when the client
// connection close unexpectedly.
func CloseHandler(h HandlerC) HandlerC {
	return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

		h.ServeHTTPC(ctx, w, r)
	})
}

// TimeoutHandler returns a Handler which adds a timeout to the context.
//
// Child handlers have the responsability to obey the context deadline and to return
// an appropriate error (or not) response in case of timeout.
func TimeoutHandler(h HandlerC, timeout time.Duration) HandlerC {
	return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx, _ = context.WithTimeout(ctx, timeout)
		h.ServeHTTPC(ctx, w, r)
	})
}
