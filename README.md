# XHandler

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/xhandler) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/xhandler/master/LICENSE) [![Build Status](https://travis-ci.org/rs/xhandler.svg?branch=master)](https://travis-ci.org/rs/xhandler) [![Coverage](http://gocover.io/_badge/github.com/rs/xhandler)](http://gocover.io/github.com/rs/xhandler)

XHandler is a bridge between [net/context](https://godoc.org/golang.org/x/net/context) and `http.Handler`.

It lets you enforce `net/context` in your handlers without sacrificing compatibility with existing `http.Handlers` nor imposing a specific router.

Thanks to `net/context` deadline management, `xhandler` is able to enforce a per request deadline and will cancel the context when the client closes the connection unexpectedly.

You may create your own `net/context` aware handler pretty much the same way as you would do with http.Handler.

Read more about xhandler on [Dailymotion engineering blog](http://engineering.dailymotion.com/our-way-to-go/).

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

func (h myMiddleware) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, "test", "World")
	h.next.ServeHTTPC(ctx, w, r)
}

func main() {
	c := xhandler.Chain{}

	// Add close notifier handler so context is cancelled when the client closes
	// the connection
	c.UseC(xhandler.CloseHandler)

	// Add timeout handler
	c.UseC(xhandler.TimeoutHandler(2 * time.Second))

	// Middleware putting something in the context
	c.UseC(func(next xhandler.HandlerC) xhandler.HandlerC {
		return myMiddleware{next: next}
	})

	// Mix it with a non-context-aware middleware handler
	c.Use(cors.Default().Handler)

	// Final handler (using handlerFuncC), reading from the context
	xh := xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		value := ctx.Value("test").(string)
		w.Write([]byte("Hello " + value))
	})

	// Bridge context aware handlers with http.Handler using xhandler.Handle()
	http.Handle("/test", c.Handler(xh))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```

### Using muxer

Xhandler comes with a context aware muxer forked from [httprouter](https://github.com/julienschmidt/httprouter):

```go
func main() {
    c := xhandler.Chain{}

	// Append a context-aware middleware handler
	c.UseC(xhandler.CloseHandler)

	// Mix it with a non-context-aware middleware handler
	c.Use(cors.Default().Handler)

	// Another context-aware middleware handler
	c.UseC(xhandler.TimeoutHandler(2 * time.Second))

	mux := xhandler.NewMux()

	// Use c.Handler to terminate the chain with your final handler
	mux.GET("/welcome/:name", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome %s!", xhandler.URLParams(ctx).Get("name"))
	}))

	if err := http.ListenAndServe(":8080", c.Handler(mux)); err != nil {
		log.Fatal(err)
	}
}
```

#### Benchmarks

Using [Julien Schmidt](https://github.com/julienschmidt) excellent [HTTP routing benchmark](https://github.com/julienschmidt/go-http-routing-benchmark), we can see that xhandler's muxer is pretty close to `httprouter` as it is a fork of it. The small overhead is due to the `net/context` allocation used to store route parameters. It still outperform other routers, thanks to amazing `httprouter`'s radix tree based matcher.

```
BenchmarkXhandler_APIStatic-8   	50000000	        39.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkChi_APIStatic-8        	 3000000	       439 ns/op	     144 B/op	       5 allocs/op
BenchmarkGoji_APIStatic-8       	 5000000	       272 ns/op	       0 B/op	       0 allocs/op
BenchmarkHTTPRouter_APIStatic-8 	50000000	        37.3 ns/op	       0 B/op	       0 allocs/op

BenchmarkXhandler_APIParam-8    	 5000000	       328 ns/op	     160 B/op	       4 allocs/op
BenchmarkChi_APIParam-8         	 2000000	       675 ns/op	     432 B/op	       6 allocs/op
BenchmarkGoji_APIParam-8        	 2000000	       692 ns/op	     336 B/op	       2 allocs/op
BenchmarkHTTPRouter_APIParam-8  	10000000	       166 ns/op	      64 B/op	       1 allocs/op

BenchmarkXhandler_API2Params-8  	 5000000	       362 ns/op	     160 B/op	       4 allocs/op
BenchmarkChi_API2Params-8       	 2000000	       814 ns/op	     432 B/op	       6 allocs/op
BenchmarkGoji_API2Params-8      	 2000000	       680 ns/op	     336 B/op	       2 allocs/op
BenchmarkHTTPRouter_API2Params-8	10000000	       183 ns/op	      64 B/op	       1 allocs/op

BenchmarkXhandler_APIAll-8      	  200000	      6473 ns/op	    2176 B/op	      64 allocs/op
BenchmarkChi_APIAll-8           	  100000	     17261 ns/op	    8352 B/op	     146 allocs/op
BenchmarkGoji_APIAll-8          	  100000	     15052 ns/op	    5377 B/op	      32 allocs/op
BenchmarkHTTPRouter_APIAll-8    	  500000	      3716 ns/op	     640 B/op	      16 allocs/op

BenchmarkXhandler_Param1-8      	 5000000	       271 ns/op	     128 B/op	       4 allocs/op
BenchmarkChi_Param1-8           	 2000000	       620 ns/op	     432 B/op	       6 allocs/op
BenchmarkGoji_Param1-8          	 3000000	       522 ns/op	     336 B/op	       2 allocs/op
BenchmarkHTTPRouter_Param1-8    	20000000	       112 ns/op	      32 B/op	       1 allocs/op

BenchmarkXhandler_Param5-8      	 3000000	       414 ns/op	     256 B/op	       4 allocs/op
BenchmarkChi_Param5-8           	 1000000	      1204 ns/op	     432 B/op	       6 allocs/op
BenchmarkGoji_Param5-8          	 2000000	       847 ns/op	     336 B/op	       2 allocs/op
BenchmarkHTTPRouter_Param5-8    	 5000000	       247 ns/op	     160 B/op	       1 allocs/op

BenchmarkXhandler_Param20-8     	 2000000	       747 ns/op	     736 B/op	       4 allocs/op
BenchmarkChi_Param20-8          	 2000000	       746 ns/op	     736 B/op	       4 allocs/op
BenchmarkGoji_Param20-8         	  500000	      2439 ns/op	    1247 B/op	       2 allocs/op
BenchmarkHTTPRouter_Param20-8   	 3000000	       585 ns/op	     640 B/op	       1 allocs/op

BenchmarkXhandler_ParamWrite-8  	 5000000	       404 ns/op	     144 B/op	       5 allocs/op
BenchmarkChi_ParamWrite-8       	 3000000	       407 ns/op	     144 B/op	       5 allocs/op
BenchmarkGoji_ParamWrite-8      	 2000000	       594 ns/op	     336 B/op	       2 allocs/op
BenchmarkHTTPRouter_ParamWrite-8	10000000	       166 ns/op	      32 B/op	       1 allocs/op
```

You can run the benchmark using `go test -bench=.` in `xhandler`'s root.

## Context Aware Middleware

Here is a list of `net/context` aware middleware handlers implementing `xhandler.HandlerC` interface.

Feel free to put up a PR linking your middleware if you have built one:

| Middleware | Author | Description |
| ---------- | ------ | ----------- |
| [xlog](https://github.com/rs/xlog) | [Olivier Poitrey](https://github.com/rs) | HTTP handler logger |
| [xstats](https://github.com/rs/xstats) | [Olivier Poitrey](https://github.com/rs) | A generic client for service instrumentation |
| [xaccess](https://github.com/rs/xaccess) | [Olivier Poitrey](https://github.com/rs) | HTTP handler access logger with [xlog](https://github.com/rs/xlog) and [xstats](https://github.com/rs/xstats) |
| [cors](https://github.com/rs/cors) | [Olivier Poitrey](https://github.com/rs) | [Cross Origin Resource Sharing](http://www.w3.org/TR/cors/) (CORS) support |

## Licenses

All source code is licensed under the [MIT License](https://raw.github.com/rs/xhandler/master/LICENSE).

Muxer is forked from [httprouter](https://github.com/julienschmidt/httprouter) with [BSD License](https://github.com/julienschmidt/httprouter/blob/master/LICENSE).
