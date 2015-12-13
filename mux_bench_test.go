// Forked from https://github.com/julienschmidt/go-http-routing-benchmark
//
package xhandler

import (
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/pressly/chi"
	goji "github.com/zenazn/goji/web"
	"golang.org/x/net/context"
)

var benchRe *regexp.Regexp

type route struct {
	method string
	path   string
}

var httpHandlerC = HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

var xhandlerWrite = HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, URLParams(ctx).Get("name"))
})

func loadXhandler(routes []route) HandlerC {
	h := namedHandler{}
	mux := NewMux()
	for _, route := range routes {
		mux.Handle(route.method, route.path, h)
	}
	return mux
}

func loadXhandlerSingle(method, path string, h HandlerC) HandlerC {
	mux := NewMux()
	mux.Handle(method, path, h)
	return mux
}

func loadChi(routes []route) HandlerC {
	h := namedHandler{}
	router := chi.NewRouter()
	for _, route := range routes {
		switch route.method {
		case "GET":
			router.Get(route.path, h)
		case "POST":
			router.Post(route.path, h)
		case "PUT":
			router.Put(route.path, h)
		case "PATCH":
			router.Patch(route.path, h)
		case "DELETE":
			router.Delete(route.path, h)
		default:
			panic("Unknown HTTP method: " + route.method)
		}
	}
	return router
}

func loadChiSingle(method, path string, h HandlerC) HandlerC {
	router := chi.NewRouter()
	switch method {
	case "GET":
		router.Get(path, h)
	case "POST":
		router.Post(path, h)
	case "PUT":
		router.Put(path, h)
	case "PATCH":
		router.Patch(path, h)
	case "DELETE":
		router.Delete(path, h)
	default:
		panic("Unknown HTTP method: " + method)
	}
	return router
}

func httpHandlerFunc(w http.ResponseWriter, r *http.Request) {}

func gojiFuncWrite(c goji.C, w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, c.URLParams["name"])
}

func loadGoji(routes []route) http.Handler {
	h := httpHandlerFunc

	mux := goji.New()
	for _, route := range routes {
		switch route.method {
		case "GET":
			mux.Get(route.path, h)
		case "POST":
			mux.Post(route.path, h)
		case "PUT":
			mux.Put(route.path, h)
		case "PATCH":
			mux.Patch(route.path, h)
		case "DELETE":
			mux.Delete(route.path, h)
		default:
			panic("Unknown HTTP method: " + route.method)
		}
	}
	return mux
}

func loadGojiSingle(method, path string, handler interface{}) http.Handler {
	mux := goji.New()
	switch method {
	case "GET":
		mux.Get(path, handler)
	case "POST":
		mux.Post(path, handler)
	case "PUT":
		mux.Put(path, handler)
	case "PATCH":
		mux.Patch(path, handler)
	case "DELETE":
		mux.Delete(path, handler)
	default:
		panic("Unknow HTTP method: " + method)
	}
	return mux
}

func httpRouterHandle(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {}

func httpRouterHandleWrite(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	io.WriteString(w, ps.ByName("name"))
}

func loadHTTPRouter(routes []route) http.Handler {
	h := httpRouterHandle

	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.method, route.path, h)
	}
	return router
}

func loadHTTPRouterSingle(method, path string, handle httprouter.Handle) http.Handler {
	router := httprouter.New()
	router.Handle(method, path, handle)
	return router
}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}

func benchRequestC(b *testing.B, router HandlerC, ctx context.Context, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTPC(ctx, w, r)
	}
}

func benchRoutes(b *testing.B, router http.Handler, routes []route) {
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.ServeHTTP(w, r)
		}
	}
}

func benchRoutesC(b *testing.B, router HandlerC, ctx context.Context, routes []route) {
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.ServeHTTPC(ctx, w, r)
		}
	}
}

// Micro Benchmarks

// Route with Param (no write)
func BenchmarkXhandler_Param1(b *testing.B) {
	router := loadXhandlerSingle("GET", "/user/:name", httpHandlerC)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkChi_Param1(b *testing.B) {
	router := loadChiSingle("GET", "/user/:name", httpHandlerC)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_Param1(b *testing.B) {
	router := loadGojiSingle("GET", "/user/:name", httpHandlerFunc)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_Param1(b *testing.B) {
	router := loadHTTPRouterSingle("GET", "/user/:name", httpRouterHandle)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}

// Route with 5 Params (no write)
const fiveColon = "/:a/:b/:c/:d/:e"
const fiveRoute = "/test/test/test/test/test"

func BenchmarkXhandler_Param5(b *testing.B) {
	router := loadXhandlerSingle("GET", fiveColon, httpHandlerC)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkChi_Param5(b *testing.B) {
	router := loadChiSingle("GET", fiveColon, httpHandlerC)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_Param5(b *testing.B) {
	router := loadGojiSingle("GET", fiveColon, httpHandlerFunc)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_Param5(b *testing.B) {
	router := loadHTTPRouterSingle("GET", fiveColon, httpRouterHandle)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequest(b, router, r)
}

// Route with 20 Params (no write)
const twentyColon = "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t"
const twentyRoute = "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t"

func BenchmarkXhandler_Param20(b *testing.B) {
	router := loadXhandlerSingle("GET", twentyColon, httpHandlerC)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkChi_Param20(b *testing.B) {
	router := loadXhandlerSingle("GET", twentyColon, httpHandlerC)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_Param20(b *testing.B) {
	router := loadGojiSingle("GET", twentyColon, httpHandlerFunc)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_Param20(b *testing.B) {
	router := loadHTTPRouterSingle("GET", twentyColon, httpRouterHandle)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequest(b, router, r)
}

// Route with Param and write
func BenchmarkXhandler_ParamWrite(b *testing.B) {
	router := loadXhandlerSingle("GET", "/user/:name", xhandlerWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkChi_ParamWrite(b *testing.B) {
	router := loadXhandlerSingle("GET", "/user/:name", xhandlerWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_ParamWrite(b *testing.B) {
	router := loadGojiSingle("GET", "/user/:name", gojiFuncWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_ParamWrite(b *testing.B) {
	router := loadHTTPRouterSingle("GET", "/user/:name", httpRouterHandleWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
