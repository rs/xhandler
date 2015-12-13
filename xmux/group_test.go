package xmux

import (
	"net/http"
	"testing"

	"github.com/rs/xhandler"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestRouteGroupOfARouteGroup(t *testing.T) {
	var get bool
	mux := New()
	foo := mux.NewGroup("/foo") // creates /foo group
	bar := foo.NewGroup("/bar")

	bar.GET("/GET", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		get = true
	}))

	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/foo/bar/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, get, "routing GET /foo/bar/GET failed")
}

func TestRouteNewGroupStripTrailingSlash(t *testing.T) {
	var get bool
	mux := New()
	foo := mux.NewGroup("/foo/")

	foo.GET("/GET", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		get = true
	}))

	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/foo/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, get, "routing GET /foo/GET failed")
}

func TestRouteNewGroupError(t *testing.T) {
	mux := New()
	assert.Panics(t, func() {
		mux.NewGroup("foo")
	})
	assert.Panics(t, func() {
		mux.NewGroup("/foo").NewGroup("bar")
	})
}

func TestRouteGroupAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete bool

	mux := New()
	group := mux.NewGroup("/foo") // creates /foo group

	group.GET("/GET", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		get = true
	}))
	group.HEAD("/GET", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		head = true
	}))
	group.OPTIONS("/GET", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		options = true
	}))
	group.POST("/POST", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		post = true
	}))
	group.PUT("/PUT", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		put = true
	}))
	group.PATCH("/PATCH", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		patch = true
	}))
	group.DELETE("/DELETE", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		delete = true
	}))

	w := new(mockResponseWriter)

	r, _ := http.NewRequest("GET", "/foo/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, get, "routing /foo/GET failed")

	r, _ = http.NewRequest("HEAD", "/foo/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, head, "routing /foo/GET failed")

	r, _ = http.NewRequest("OPTIONS", "/foo/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, options, "routing /foo/GET failed")

	r, _ = http.NewRequest("POST", "/foo/POST", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, post, "routing /foo/POST failed")

	r, _ = http.NewRequest("PUT", "/foo/PUT", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, put, "routing /foo/PUT failed")

	r, _ = http.NewRequest("PATCH", "/foo/PATCH", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, patch, "routing /foo/PATCH failed")

	r, _ = http.NewRequest("DELETE", "/foo/DELETE", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, delete, "routing /foo/DELETE failed")
}
