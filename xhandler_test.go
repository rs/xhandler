package xhandler

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type handler struct{}
type key int

const contextKey key = 0

func newContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, contextKey, value)
}

func fromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(contextKey).(string)
	return value, ok
}

func (h handler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	value, _ := fromContext(ctx)
	if _, ok := ctx.Deadline(); ok {
		value += " with deadline"
	}
	w.Write([]byte(value))
}

func TestHandle(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "value")
	h := Handle(ctx, &handler{})
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, "value", w.Body.String())
}

func TestHandleTimeout(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "value")
	h := HandleTimeout(ctx, time.Second, &handler{})
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, "value with deadline", w.Body.String())
}

func TestCtxHandlerFunc(t *testing.T) {
	ok := false
	xh := CtxHandlerFunc(func(context.Context, http.ResponseWriter, *http.Request) {
		ok = true
	})
	xh.ServeHTTP(nil, nil, nil)
	assert.True(t, ok)
}
