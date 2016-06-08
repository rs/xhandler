package xhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendHandlerC(t *testing.T) {
	init := 0
	h1 := func(next http.Handler) http.Handler {
		init++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "test", 1)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	h2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "test", 2)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	c := Chain{}
	c.Use(h1)
	c.Use(h2)
	assert.Len(t, c, 2)

	h := c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test ordering
		assert.Equal(t, 2, r.Context().Value("test"), "second handler should overwrite first handler's context value")
	}))

	req := httptest.NewRequest("", "/test", nil)

	h.ServeHTTP(nil, req)
	h.ServeHTTP(nil, req)

	assert.Equal(t, 1, init, "handler init called once")
}

func TestAppendHandler(t *testing.T) {
	init := 0
	h1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "test", 1)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	h2 := func(next http.Handler) http.Handler {
		init++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Change r and w values
			w = httptest.NewRecorder()
			// TODO: since the context lives inside the context now, this wouldn't be fair (meaning: work)
			//r = &http.Request{}
			next.ServeHTTP(w, r)
		})
	}
	c := Chain{}
	c.Use(h1)
	c.Use(h2)
	assert.Len(t, c, 2)

	h := c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test ordering
		assert.Equal(t, 1, r.Context().Value("test"),
			"the first handler value should be pass through the second (non-aware) one")
		// Test r and w overwrite
		assert.NotNil(t, w)
		assert.NotNil(t, r)
	}))

	req := httptest.NewRequest("", "/test", nil)

	h.ServeHTTP(nil, req)
	h.ServeHTTP(nil, req)

	// There's no safe way to not initialize non ctx aware handlers on each request :/
	//assert.Equal(t, 1, init, "handler init called once")
}

func TestChainHandlerC(t *testing.T) {
	handlerCalls := 0
	h1 := func(next http.Handler) http.Handler {
		handlerCalls++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "test", 1)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	h2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalls++
			ctx := context.WithValue(r.Context(), "test", 2)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	c := Chain{}
	c.Use(h1)
	c.Use(h2)
	h := c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalls++

		assert.Equal(t, 2, r.Context().Value("test"),
			"second handler should overwrite first handler's context value")
		assert.Equal(t, 1, r.Context().Value("mainCtx"),
			"the mainCtx value should be pass through")
	}))

	req := httptest.NewRequest("", "/test", nil)
	mainCtx := context.WithValue(context.Background(), "mainCtx", 1)
	h.ServeHTTP(nil, req.WithContext(mainCtx))

	assert.Equal(t, 3, handlerCalls, "all handler called once")
}
