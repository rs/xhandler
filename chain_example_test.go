package xhandler_test

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/xhandler"
)

func ExampleChain() {
	c := xhandler.Chain{}
	// Append a context-aware middleware handler
	c.Use(xhandler.CloseHandler)

	// Mix it with a non-context-aware middleware handler
	// TODO: adapt new api for cors
	//c.Use(cors.Default().Handler)

	// Another context-aware middleware handler
	c.Use(xhandler.TimeoutHandler(2 * time.Second))

	mux := http.NewServeMux()

	// Use c.Handler to terminate the chain with your final handler
	mux.Handle("/", c.Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})))

	// You can reuse the same chain for other handlers
	mux.Handle("/api", c.Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the API!")
	})))
}

func ExampleIf() {
	c := xhandler.Chain{}

	// Add timeout handler only if the path match a prefix
	c.Use(xhandler.If(
		func(w http.ResponseWriter, r *http.Request) bool {
			return strings.HasPrefix(r.URL.Path, "/with-timeout/")
		},
		xhandler.TimeoutHandler(2*time.Second),
	))

	http.Handle("/", c.Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})))
}
