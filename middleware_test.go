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

func TestTimeoutHandler(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "value")
	xh := TimeoutHandler(time.Second)(&handler{})
	h := New(ctx, xh)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, "value with deadline", w.Body.String())
}

func TestBasicAuthHandler(t *testing.T) {
	ctx := context.Background()
	xh := BasicAuthHandler("user", "pass")(&handler{})
	h := New(ctx, xh)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)

	w = httptest.NewRecorder()
	r, err = http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	r.SetBasicAuth("user", "pass")
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

type closeNotifyWriter struct {
	*httptest.ResponseRecorder
}

func (w *closeNotifyWriter) CloseNotify() <-chan bool {
	// return an already "closed" notifier
	notify := make(chan bool, 1)
	notify <- true
	return notify
}

func TestCloseHandler(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "value")
	xh := CloseHandler(&handler{})
	h := New(ctx, xh)
	w := &closeNotifyWriter{httptest.NewRecorder()}
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, "value canceled", w.Body.String())
}

func TestIf(t *testing.T) {
	trueHandler := HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/true", r.URL.Path)
	})
	falseHandler := HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, "/true", r.URL.Path)
	})
	ctx := context.WithValue(context.Background(), contextKey, "value")
	xh := If(
		func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			return r.URL.Path == "/true"
		},
		func(next HandlerC) HandlerC {
			return trueHandler
		},
	)(falseHandler)
	h := New(ctx, xh)
	r, _ := http.NewRequest("GET", "http://example.com/true", nil)
	h.ServeHTTP(nil, r)
	r, _ = http.NewRequest("GET", "http://example.com/false", nil)
	h.ServeHTTP(nil, r)
}
