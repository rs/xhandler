package xhandler_test

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rs/xhandler"
)

type key int

const contextKey key = 0

func newContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, contextKey, value)
}

func fromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(contextKey).(string)
	return value, ok
}

func ExampleHandle() {
	var xh http.Handler
	xh = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value, _ := fromContext(r.Context())
		w.Write([]byte("Hello " + value))
	})

	xh = (func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := newContext(r.Context(), "World")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})(xh)

	ctx := context.Background()
	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	http.Handle("/", xhandler.New(ctx, xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func ExampleHandleTimeout() {
	var xh http.Handler
	xh = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
		if _, ok := r.Context().Deadline(); ok {
			w.Write([]byte(" with deadline"))
		}
	})

	// This handler adds a timeout to the handler
	xh = xhandler.TimeoutHandler(5 * time.Second)(xh)

	ctx := context.Background()
	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	http.Handle("/", xhandler.New(ctx, xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
