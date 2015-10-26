# XHandler

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/xhandler) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/xhandler/master/LICENSE) [![Build Status](https://travis-ci.org/rs/xhandler.svg?branch=master)](https://travis-ci.org/rs/xhandler) [![Coverage](http://gocover.io/_badge/github.com/rs/xhandler)](http://gocover.io/github.com/rs/xhandler)

XHandler is a bridge between [net/context](https://godoc.org/golang.org/x/net/context) and `http.Handler`.

It lets you enforce `net/context` in your handlers without sacrificing compatibility with existing `http.Handlers` nor imposing a specific router.

Thanks to `net/context` deadline management, `xhandler` is able to enforce a per request deadline and will cancel the context when the client closes the connection unexpectedly.

You may create your own `net/context` aware handler pretty much the same way as you would do with http.Handler.

This library is inspired by https://joeshaw.org/net-context-and-http-handler/.

## Installing

    go get -u github.com/rs/xhandler

## Usage

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

type myMiddleware struct {
	next xhandler.HandlerC
}

func (h *myMiddleware) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = newContext(ctx, "World")
	h.next.ServeHTTPC(ctx, w, r)
}

func main() {
	var xh xhandler.HandlerC

	// Inner handler (using handler func), reading from the context
	xh = xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		value, _ := fromContext(ctx)
		w.Write([]byte("Hello " + value))
	})

	// Middleware putting something in the context
	xh = &myMiddleware{next: xh}

	// Add close notifier handler so context is cancelled when the client closes
	// the connection
	xh = xhandler.CloseHandler(xh)

	// Add timeout handler
	xh = xhandler.TimeoutHandler(xh, 5*time.Second)

	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	// This handler is now a conventional http.Handler which can be wrapped
	// by any other non context aware handlers.
	var h http.Handler
	// Root context
	ctx := context.Background()
	h = xhandler.New(ctx, xh)

	// As an example, we wrap this handler into the non context aware CORS handler
	h = cors.Default().Handler(h)

	http.Handle("/", h)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```

How to use it with a non- `net/context` aware router. Lets try with the Go's `ServerMux`:

```go
func main() {
    mux := http.NewServeMux()
    // Use xwrap to insert xhandler and all the intermediate context handlers
    mux.Handle("/api/", xwrap(apiHandlerC{}))
    mux.HandleFunc("/", xwrap(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
        fmt.Fprintf(w, "Welcome to the home page!")
    }))
}

func xwrap(xh xhandler.HandlerC) http.Handler {
    // Middleware putting something in the context
    xh = &myMiddleware{next: xh}

    ctx := context.Background()
    return xhandler.New(ctx, xh)
}
```

## Context Aware Middleware

Here is a list of `net/context` aware middleware handlers implementing `xhandler.HandlerC` interface.

Feel free to put up a PR linking your middleware if you have built one:

| Middleware | Author | Description |
| ---------- | ------ | ----------- |
| [xlog](https://github.com/rs/xlog) | [Olivier Poitrey](https://github.com/rs) | HTTP handler logger |
| [xstats](https://github.com/rs/xstats) | [Olivier Poitrey](https://github.com/rs) | A generic client for service instrumentation |
| [cors](https://github.com/rs/cors) | [Olivier Poitrey](https://github.com/rs) | [Cross Origin Resource Sharing](http://www.w3.org/TR/cors/) (CORS) support |

## Licenses

All source code is licensed under the [MIT License](https://raw.github.com/rs/xhandler/master/LICENSE).
