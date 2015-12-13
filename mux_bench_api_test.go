// Forked from https://github.com/julienschmidt/go-http-routing-benchmark
//
package xhandler

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"
)

// Parse API
// https://parse.com/docs/rest#summary
var parseAPI = []route{
	// Objects
	{"POST", "/1/classes/:className"},
	{"GET", "/1/classes/:className/:objectId"},
	{"PUT", "/1/classes/:className/:objectId"},
	{"GET", "/1/classes/:className"},
	{"DELETE", "/1/classes/:className/:objectId"},

	// Users
	{"POST", "/1/users"},
	{"GET", "/1/login"},
	{"GET", "/1/users/:objectId"},
	{"PUT", "/1/users/:objectId"},
	{"GET", "/1/users"},
	{"DELETE", "/1/users/:objectId"},
	{"POST", "/1/ruestPasswordReset"},

	// Roles
	{"POST", "/1/roles"},
	{"GET", "/1/roles/:objectId"},
	{"PUT", "/1/roles/:objectId"},
	{"GET", "/1/roles"},
	{"DELETE", "/1/roles/:objectId"},

	// Files
	{"POST", "/1/files/:fileName"},

	// Analytics
	{"POST", "/1/events/:eventName"},

	// Push Notifications
	{"POST", "/1/push"},

	// Installations
	{"POST", "/1/installations"},
	{"GET", "/1/installations/:objectId"},
	{"PUT", "/1/installations/:objectId"},
	{"GET", "/1/installations"},
	{"DELETE", "/1/installations/:objectId"},

	// Cloud Functions
	{"POST", "/1/functions"},
}

var (
	parseXhandler   HandlerC
	parseChi        HandlerC
	parseGoji       http.Handler
	parseHTTPRouter http.Handler
)

func getParseXhandler(b *testing.B) HandlerC {
	defer b.ResetTimer()
	if parseXhandler == nil {
		parseXhandler = loadXhandler(parseAPI)
	}
	return parseXhandler
}

func getParseChi(b *testing.B) HandlerC {
	defer b.ResetTimer()
	if parseChi == nil {
		parseChi = loadChi(parseAPI)
	}
	return parseChi
}

func getParseGoji(b *testing.B) http.Handler {
	defer b.ResetTimer()
	if parseGoji == nil {
		parseGoji = loadGoji(parseAPI)
	}
	return parseGoji
}

func getParseHTTPRouter(b *testing.B) http.Handler {
	defer b.ResetTimer()
	if parseHTTPRouter == nil {
		parseHTTPRouter = loadHTTPRouter(parseAPI)
	}
	return parseHTTPRouter
}

func BenchmarkXhandler_APIStatic(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequestC(b, getParseXhandler(b), context.Background(), r)
}
func BenchmarkChi_APIStatic(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequestC(b, getParseChi(b), context.Background(), r)
}
func BenchmarkGoji_APIStatic(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, getParseGoji(b), r)
}
func BenchmarkHTTPRouter_APIStatic(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, getParseHTTPRouter(b), r)
}

func BenchmarkXhandler_APIParam(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequestC(b, getParseXhandler(b), context.Background(), r)
}
func BenchmarkChi_APIParam(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequestC(b, getParseChi(b), context.Background(), r)
}
func BenchmarkGoji_APIParam(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, getParseGoji(b), r)
}
func BenchmarkHTTPRouter_APIParam(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, getParseHTTPRouter(b), r)
}

func BenchmarkXhandler_API2Params(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequestC(b, getParseXhandler(b), context.Background(), r)
}
func BenchmarkChi_API2Params(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequestC(b, getParseChi(b), context.Background(), r)
}
func BenchmarkGoji_API2Params(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, getParseGoji(b), r)
}
func BenchmarkHTTPRouter_API2Params(b *testing.B) {
	r, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, getParseHTTPRouter(b), r)
}

func BenchmarkXhandler_APIAll(b *testing.B) {
	benchRoutesC(b, getParseXhandler(b), context.Background(), parseAPI)
}
func BenchmarkChi_APIAll(b *testing.B) {
	benchRoutesC(b, getParseChi(b), context.Background(), parseAPI)
}
func BenchmarkGoji_APIAll(b *testing.B) {
	benchRoutes(b, getParseGoji(b), parseAPI)
}
func BenchmarkHTTPRouter_APIAll(b *testing.B) {
	benchRoutes(b, getParseHTTPRouter(b), parseAPI)
}
