package xhandler_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/cors"
	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

func ExampleChain() {
	c := xhandler.Chain{}
	// Append a context-aware middleware handler
	c.UseC(xhandler.CloseHandler)

	// Mix it with a non-context-aware middleware handler
	c.Use(cors.Default().Handler)

	// Another context-aware middleware handler
	c.UseC(xhandler.TimeoutHandler(2 * time.Second))

	mux := http.NewServeMux()

	// Use c.Handler to terminate the chain with your final handler
	mux.Handle("/", c.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})))

	// You can reuse the same chain for other handlers
	mux.Handle("/api", c.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the API!")
	})))
}
