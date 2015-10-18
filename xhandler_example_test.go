package xhandler_test

import (
	"log"
	"net/http"
	"time"

	"github.com/rs/xhandler"
	"golang.org/x/net/context"
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
	var xh xhandler.Handler
	xh = xhandler.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		value, _ := fromContext(ctx)
		w.Write([]byte("Hello " + value))
	})

	xh = (func(next xhandler.Handler) xhandler.Handler {
		return xhandler.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = newContext(ctx, "World")
			next.ServeHTTP(ctx, w, r)
		})
	})(xh)

	ctx := context.Background()
	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	http.Handle("/", xhandler.CtxHandler(ctx, xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func ExampleHandleTimeout() {
	var xh xhandler.Handler
	xh = xhandler.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
		if _, ok := ctx.Deadline(); ok {
			w.Write([]byte(" with deadline"))
		}
	})

	// This handler adds a timeout to the handler
	xh = xhandler.TimeoutHandler(xh, 5*time.Second)

	ctx := context.Background()
	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	http.Handle("/", xhandler.CtxHandler(ctx, xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
