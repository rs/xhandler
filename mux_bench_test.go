package xhandler

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	goji "github.com/zenazn/goji/web"
	"golang.org/x/net/context"
)

var benchRe *regexp.Regexp

type route struct {
	method string
	path   string
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func isTested(name string) bool {
	if benchRe == nil {
		// Get -test.bench flag value (not accessible via flag package)
		bench := ""
		for _, arg := range os.Args {
			if strings.HasPrefix(arg, "-test.bench=") {
				// ignore the benchmark name after an underscore
				bench = strings.SplitN(arg[12:], "_", 2)[0]
				break
			}
		}

		// Compile RegExp to match Benchmark names
		var err error
		benchRe, err = regexp.Compile(bench)
		if err != nil {
			panic(err.Error())
		}
	}
	return benchRe.MatchString(name)
}

func calcMem(name string, load func()) {
	if !isTested(name) {
		return
	}

	m := new(runtime.MemStats)

	// before
	runtime.GC()
	runtime.ReadMemStats(m)
	before := m.HeapAlloc

	load()

	// after
	runtime.GC()
	runtime.ReadMemStats(m)
	after := m.HeapAlloc
	println("   "+name+":", after-before, "Bytes")
}

var httpHandlerC = HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

var xhandlerWrite = HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, URLParams(ctx)["name"])
})

func loadXhandler(routes []route) HandlerC {
	mux := NewMux()
	for _, route := range routes {
		mux.Handle(route.method, route.path, namedHandler{})
	}
	return mux
}

func loadXhandlerSingle(method, path string, handle HandlerC) HandlerC {
	mux := NewMux()
	mux.Handle(method, path, handle)
	return mux
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
func BenchmarkXhandler_Param(b *testing.B) {
	router := loadXhandlerSingle("GET", "/user/:name", httpHandlerC)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_Param(b *testing.B) {
	router := loadGojiSingle("GET", "/user/:name", httpHandlerFunc)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_Param(b *testing.B) {
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
