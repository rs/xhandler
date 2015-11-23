package xhandler

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// CloseHandler returns a Handler cancelling the context when the client
// connection close unexpectedly.
func CloseHandler(next HandlerC) HandlerC {
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

		next.ServeHTTPC(ctx, w, r)
	})
}

// TimeoutHandler returns a Handler which adds a timeout to the context.
//
// Child handlers have the responsability to obey the context deadline and to return
// an appropriate error (or not) response in case of timeout.
func TimeoutHandler(timeout time.Duration) func(next HandlerC) HandlerC {
	return func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx, _ = context.WithTimeout(ctx, timeout)
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// BasicAuthHandler returns a handler which errors out with
// http.StatusForbidden if the correct credentials were not provided in the
// request.
func BasicAuthHandler(user, pass string) func(next HandlerC) HandlerC {
	return func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if usr, pw, ok := r.BasicAuth(); !ok || usr != user || pw != pass {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// If is a special handler that will skip insert the condNext handler only if a condition
// applies at runtime.
func If(cond func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, condNext func(next HandlerC) HandlerC) func(next HandlerC) HandlerC {
	return func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if cond(ctx, w, r) {
				condNext(next).ServeHTTPC(ctx, w, r)
			} else {
				next.ServeHTTPC(ctx, w, r)
			}
		})
	}
}
