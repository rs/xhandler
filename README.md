# XHandler

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/xhandler) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/xhandler/master/LICENSE) [![Build Status](https://travis-ci.org/rs/xhandler.svg?branch=master)](https://travis-ci.org/rs/xhandler)

XHandler is a bridge between [net/context](https://godoc.org/golang.org/x/net/context) and `http.Handler`.

It lets you enforce `net/context` in your handlers without sacrificing compatibility with existing `http.Handlers` nor imposing a specific router.

Thanks to `net/context` deadline management, `xhandler` is able to enforce a per request deadline and will cancel the context when the client closes the connection unexpectedly.

You may create your own `net/context` aware middlewares pretty much the same way as you would do with http.Handler.

## Installing

    go get -u github.com/rs/xhandler

## Example

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

type myMiddleware struct {
	next xhandler.CtxHandler
}

func (h *myMiddleware) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = newContext(ctx, "World")
	h.next.ServeHTTP(ctx, w, r)
}

func main() {
	var xh xhandler.CtxHandler

	// Inner handler (using handler func), reading from the context
	xh = xhandler.CtxHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		value, _ := fromContext(ctx)
		w.Write([]byte("Hello " + value))
	})

	// Middleware putting something in the context
	xh = &myMiddleware{next: xh}

	// Root context
	ctx := context.Background()
	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	// Use HandleTimeout() if you want to set a per request timeout.
	http.Handle("/", xhandler.Handle(ctx, xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```

## Licenses

All source code is licensed under the [MIT License](https://raw.github.com/rs/xhandler/master/LICENSE).
