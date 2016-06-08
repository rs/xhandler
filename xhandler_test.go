package xhandler

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Leave other go routines a chance to run
	time.Sleep(time.Nanosecond)
	value, _ := fromContext(ctx)
	if _, ok := ctx.Deadline(); ok {
		value += " with deadline"
	}
	if ctx.Err() == context.Canceled {
		value += " canceled"
	}
	w.Write([]byte(value))
}

func TestHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), contextKey, "value")
	r = r.WithContext(ctx)
	h := &handler{}

	h.ServeHTTP(w, r)
	assert.Equal(t, "value", w.Body.String())
}

func TestHandlerFunc(t *testing.T) {
	ok := false
	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		ok = true
	})
	h.ServeHTTP(nil, nil)
	assert.True(t, ok)
}
