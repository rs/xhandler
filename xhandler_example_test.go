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
	xh := xhandler.CtxHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		value, _ := fromContext(ctx)
		w.Write([]byte("Hello " + value))
	})

	xh = (func(next xhandler.CtxHandlerFunc) xhandler.CtxHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = newContext(ctx, "World")
			next(ctx, w, r)
		}
	})(xh)

	ctx := context.Background()
	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	http.Handle("/", xhandler.Handle(ctx, xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func ExampleHandleTimeout() {
	xh := xhandler.CtxHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		value, _ := fromContext(ctx)
		w.Write([]byte("Hello " + value))
	})

	xh = (func(next xhandler.CtxHandlerFunc) xhandler.CtxHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = newContext(ctx, "World")
			next(ctx, w, r)
		}
	})(xh)

	ctx := context.Background()
	// Bridge context aware handlers with http.Handler using xhandler.HandleTimout()
	// The provided timout will be set on each request's context
	http.Handle("/", xhandler.HandleTimeout(ctx, 5*time.Second, xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
