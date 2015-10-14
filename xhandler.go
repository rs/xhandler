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

// Handler is a net/context aware http handler bridge between http.Handler and net/context.
type Handler struct {
	ctx     context.Context
	timeout time.Duration
	next    CtxHandler
}

// CtxHandler is a net/context aware http.Handler
type CtxHandler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

// The CtxHandlerFunc type is an adapter to allow the use of ordinary functions
// as a CtxHandler. If f is a function with the appropriate signature,
// CtxHandlerFunc(f) is a CtxHandler object that calls f.
type CtxHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTP calls f(w, r).
func (f CtxHandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

// Handle returns a new xhandler.Handler
//
// The provided context is used as root context.
func Handle(ctx context.Context, next CtxHandler) *Handler {
	return &Handler{
		ctx:  ctx,
		next: next,
	}
}

// HandleTimeout creates a new xhandler.Handler with per request timeout
func HandleTimeout(ctx context.Context, timeout time.Duration, next CtxHandler) *Handler {
	return &Handler{
		ctx:     ctx,
		timeout: timeout,
		next:    next,
	}
}

// ServeHTTP implements http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var ctx context.Context
	var cancel context.CancelFunc
	if h.timeout > 0 {
		ctx, cancel = context.WithTimeout(h.ctx, h.timeout)
	} else {
		ctx, cancel = context.WithCancel(h.ctx)
	}

	// Cancel the context if the client closes the connection
	if wcn, ok := w.(http.CloseNotifier); ok {
		notify := wcn.CloseNotify()
		go func() {
			<-notify
			cancel()
		}()
	}

	// Call next handler
	h.next.ServeHTTP(ctx, w, r)

	cancel()
}
