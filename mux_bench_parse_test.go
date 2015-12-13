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
	{"POST", "/1/requestPasswordReset"},

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
	parseGoji       http.Handler
	parseHTTPRouter http.Handler
)

func init() {
	println("#ParseAPI Routes:", len(parseAPI))

	calcMem("Xhandler", func() {
		parseXhandler = loadXhandler(parseAPI)
	})

	calcMem("Goji", func() {
		parseGoji = loadGoji(parseAPI)
	})

	calcMem("HTTPRouter", func() {
		parseHTTPRouter = loadHTTPRouter(parseAPI)
	})

	println()
}

func BenchmarkXhandler_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequestC(b, parseXhandler, context.Background(), req)
}

func BenchmarkGoji_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseGoji, req)
}

func BenchmarkHttpRouter_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseHTTPRouter, req)
}

func BenchmarkXhandler_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequestC(b, parseXhandler, context.Background(), req)
}
func BenchmarkGoji_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseGoji, req)
}
func BenchmarkHttpRouter_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseHTTPRouter, req)
}

func BenchmarkXhandler_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequestC(b, parseXhandler, context.Background(), req)
}
func BenchmarkGoji_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseGoji, req)
}
func BenchmarkHttpRouter_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseHTTPRouter, req)
}

func BenchmarkXhandler_ParseAll(b *testing.B) {
	benchRoutesC(b, parseXhandler, context.Background(), parseAPI)
}
func BenchmarkGoji_ParseAll(b *testing.B) {
	benchRoutes(b, parseGoji, parseAPI)
}
func BenchmarkHttpRouter_ParseAll(b *testing.B) {
	benchRoutes(b, parseHTTPRouter, parseAPI)
}
