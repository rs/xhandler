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
