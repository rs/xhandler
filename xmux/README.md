# Xmux

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/xhandler/xmux) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/xhandler/master/LICENSE) [![Build Status](https://travis-ci.org/rs/xhandler.svg?branch=master)](https://travis-ci.org/rs/xhandler) [![Coverage](http://gocover.io/_badge/github.com/rs/xhandler/xmux)](http://gocover.io/github.com/rs/xhandler/xmux)

Xmux is a lightweight high performance HTTP request muxer on top [xhandler](https://github.com/rs/xhandler). Xmux gets its speed from the fork of the amazing [httprouter](https://github.com/julienschmidt/httprouter). Route parameters are stored in `net/context` instead of being passed as an additional parameter.

In contrast to the [default mux](http://golang.org/pkg/net/http/#ServeMux) of Go's `net/http` package, this muxer supports variables in the routing pattern and matches against the request method. It also scales better.

The muxer is optimized for high performance and a small memory footprint. It scales well even with very long paths and a large number of routes. A compressing dynamic trie (radix tree) structure is used for efficient matching.

## Features

**Only explicit matches:** With other muxers, like [http.ServeMux](http://golang.org/pkg/net/http/#ServeMux), a requested URL path could match multiple patterns. Therefore they have some awkward pattern priority rules, like *longest match* or *first registered, first matched*. By design of this router, a request can only match exactly one or no route. As a result, there are also no unintended matches, which makes it great for SEO and improves the user experience.

**Stop caring about trailing slashes:** Choose the URL style you like, the muxer automatically redirects the client if a trailing slash is missing or if there is one extra. Of course it only does so, if the new path has a handler. If you don't like it, you can [turn off this behavior](http://godoc.org/github.com/rs/xhandler/xmux#Mux.RedirectTrailingSlash).

**Path auto-correction:** Besides detecting the missing or additional trailing slash at no extra cost, the muxer can also fix wrong cases and remove superfluous path elements (like `../` or `//`). Is [CAPTAIN CAPS LOCK](http://www.urbandictionary.com/define.php?term=Captain+Caps+Lock) one of your users? Xmux can help him by making a case-insensitive look-up and redirecting him to the correct URL.

**Parameters in your routing pattern:** Stop parsing the requested URL path, just give the path segment a name and the router delivers the dynamic value to you. Because of the design of the router, path parameters are very cheap.

**RouteGroups:** A way to create [groups of routes](http://godoc.org/github.com/rs/xhandler/xmux#Mux.NewGroup) without incurring any per-request overhead.

**Zero Garbage:** The matching and dispatching process generates zero bytes of garbage. In fact, the only heap allocations that are made, is by building the slice of the key-value pairs for path parameters and the `net/context` instance to store them in the context. If the request path contains no parameters, not a single heap allocation is necessary.

**No more server crashes:** You can set a [Panic handler](http://godoc.org/github.com/rs/xhandler/xmux#Mux.PanicHandler) to deal with panics occurring during handling a HTTP request. The router then recovers and lets the `PanicHandler` log what happened and deliver a nice error page.

Of course you can also set **custom [NotFound](http://godoc.org/github.com/rs/xhandler/xmux#Mux.NotFound) and  [MethodNotAllowed](http://godoc.org/github.com/rs/xhandler/xmux#Mux.MethodNotAllowed) handlers**.

## Usage

This is just a quick introduction, view the [GoDoc](http://godoc.org/github.com/rs/xhandler/xmux) for details.

Let's start with a trivial example:
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rs/xhandler"
	"github.com/rs/xhandler/xmux"
	"golang.org/x/net/context"
)

func Index(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, %s!\n", xmux.Params(ctx).Get("name"))
}

func main() {
	mux := xmux.New()
	mux.GET("/", Index)
	mux.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", xhandler.New(context.Background(), mux)))
}
```

You may also chain middleware using `xhandler.Chain`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rs/xhandler"
	"github.com/rs/xhandler/xmux"
	"golang.org/x/net/context"
)

func main() {
	c := xhandler.Chain{}

	// Append a context-aware middleware handler
	c.UseC(xhandler.CloseHandler)

	// Another context-aware middleware handler
	c.UseC(xhandler.TimeoutHandler(2 * time.Second))

	mux := xmux.New()

	// Use c.Handler to terminate the chain with your final handler
	mux.GET("/welcome/:name", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome %s!", xmux.Params(ctx).Get("name"))
	}))

	if err := http.ListenAndServe(":8080", c.Handler(mux)); err != nil {
		log.Fatal(err)
	}
}
```

### Named parameters

As you can see, `:name` is a *named parameter*. The values are accessible via `xmux.Params(ctx)`, which returns `xmux.ParamHolder`.
You can get the value of a parameter by its name using `Get(name)` method:

Named parameters only match a single path segment:

```
Pattern: /user/:user

 /user/gordon              match
 /user/you                 match
 /user/gordon/profile      no match
 /user/                    no match
```

**Note:** Since this muxer has only explicit matches, you can not register static routes and parameters for the same path segment. For example you can not register the patterns `/user/new` and `/user/:user` for the same request method at the same time. The routing of different request methods is independent from each other.

### Catch-All parameters

The second type are *catch-all* parameters and have the form `*name`. Like the name suggests, they match everything. Therefore they must always be at the **end** of the pattern:

```
Pattern: /src/*filepath

 /src/                     match
 /src/somefile.go          match
 /src/subdir/somefile.go   match
```

## Benchmarks

Thanks to [Julien Schmidt](https://github.com/julienschmidt) excellent [HTTP routing benchmark](https://github.com/julienschmidt/go-http-routing-benchmark), we can see that xhandler's muxer is pretty close to `httprouter` as it is a fork of it. The small overhead is due to the `net/context` allocation used to store route parameters. It still outperform other routers, thanks to amazing `httprouter`'s radix tree based matcher.

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

You can run this benchmark by using `go test -bench=.` in `xmux`'s root.


## Licenses

All source code is licensed under the [MIT License](https://raw.github.com/rs/xhandler/master/LICENSE).

Xmux is forked from [httprouter](https://github.com/julienschmidt/httprouter) with [BSD License](https://github.com/julienschmidt/httprouter/blob/master/LICENSE).
