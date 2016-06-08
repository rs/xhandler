package xhandler

import (
	"context"
	"net/http"
	"time"
)

// CloseHandler returns a Handler cancelling the context when the client
// connection close unexpectedly.
// TODO: not sure this is also in the stdlib now
func CloseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Cancel the context if the client closes the connection
		if wcn, ok := w.(http.CloseNotifier); ok {
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			notify := wcn.CloseNotify()
			go func() {
				select {
				case <-notify:
					cancel()
				case <-ctx.Done():
				}
			}()
			// TODO: needed?
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// TimeoutHandler returns a Handler which adds a timeout to the context.
//
// Child handlers have the responsability to obey the context deadline and to return
// an appropriate error (or not) response in case of timeout.
func TimeoutHandler(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: not calling cancel?!
			ctx, _ := context.WithTimeout(r.Context(), timeout)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// If is a special handler that will skip insert the condNext handler only if a condition
// applies at runtime.
func If(cond func(w http.ResponseWriter, r *http.Request) bool, condNext func(next http.Handler) http.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cond(w, r) {
				condNext(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
