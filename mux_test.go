package xhandler

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func TestParsePath(t *testing.T) {
	assert.Equal(t, []pos{}, parsePath("/"))
	assert.Equal(t, []pos{pos{start: 1, end: 2}}, parsePath("/a"))
	assert.Equal(t, []pos{pos{start: 1, end: 2}, pos{start: 3, end: 4}}, parsePath("/a/b"))
	assert.Equal(t, []pos{pos{start: 1, end: 2}, pos{start: 3, end: 4}}, parsePath("/a/b/"))
	assert.Equal(t, []pos{pos{start: 1, end: 4}, pos{start: 5, end: 8}}, parsePath("/foo/bar"))
	assert.Equal(t, []pos{pos{start: 1, end: 4}, pos{start: 6, end: 9}}, parsePath("/foo//bar/"))
}

func BenchmarkParsePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parsePath("/test/test/test/test/test")
	}
}

func testPattern(path, pat string) bool {
	return matchPattern(path, parsePath(path), newPattern(pat))
}

func TestMatchPatternStatic(t *testing.T) {
	assert.True(t, testPattern("/a/b/c", "/a/b/c"))
	assert.True(t, testPattern("/foo/bar/baz", "/foo/bar/baz"))
	assert.False(t, testPattern("/a/b/c", "/a/b/c/d"))
	assert.False(t, testPattern("/foo/bar/baz", "/foo/bar/baz/baz"))
	assert.False(t, testPattern("/a/b/c/d", "/a/b/c"))
	assert.False(t, testPattern("/foo/bar/baz/baz", "/foo/bar/baz"))
	assert.False(t, testPattern("/a/b/c", "/a/b/d"))
	assert.True(t, testPattern("/a/b/c", "/a/b//c"))
	assert.True(t, testPattern("/a/b//c", "/a/b//c"))
	assert.True(t, testPattern("/a/b/c", "/a/b/"))
}

func BenchmarkMatchPatternStatic(b *testing.B) {
	path := "/foo/bar/baz"
	components := parsePath(path)
	pat := newPattern("/foo/bar/baz")
	for i := 0; i < b.N; i++ {
		matchPattern(path, components, pat)
	}
}

func BenchmarkMatchPatternStaticPrefix(b *testing.B) {
	path := "/foo/bar/baz"
	components := parsePath(path)
	pat := newPattern("/foo/bar/")
	for i := 0; i < b.N; i++ {
		matchPattern(path, components, pat)
	}
}

func TestMatchPatternVar(t *testing.T) {
	assert.True(t, testPattern("/a/b/c", "/a/:var/c"))
	assert.False(t, testPattern("/a/b/b/c", "/a/:var/c"))
	assert.True(t, testPattern("/a/b/c", "/:var/b/c"))
	assert.True(t, testPattern("/a", "/:var"))
	assert.True(t, testPattern("/a/b/c", "/:var/b/c"))
}

func BenchmarkMatchPatternVar(b *testing.B) {
	path := "/foo/bar/baz"
	components := parsePath(path)
	pat := newPattern("/foo/:var/baz")
	for i := 0; i < b.N; i++ {
		matchPattern(path, components, pat)
	}
}

type namedHandler struct {
	name string
}

func (n namedHandler) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
}

func TestMuxMatchStatic(t *testing.T) {
	a := namedHandler{name: "a"}
	b := namedHandler{name: "b"}
	c := namedHandler{name: "c"}
	mux := Mux{}
	mux.Handle("GET", "/foo/bar/baz", a)
	mux.Handle("GET", "/foo/bar/", b)
	mux.Handle("GET", "/", c)
	m, p := mux.match("GET", "/foo/bar/baz")
	assert.Nil(t, p)
	if assert.NotNil(t, m) {
		assert.Equal(t, a, m.h)
	}
	m, p = mux.match("GET", "/foo/bar/bar")
	assert.Nil(t, p)
	if assert.NotNil(t, m) {
		assert.Equal(t, b, m.h)
	}
	m, p = mux.match("GET", "/")
	assert.Nil(t, p)
	if assert.NotNil(t, m) {
		assert.Equal(t, c, m.h)
	}
	m, p = mux.match("GET", "/something")
	assert.Nil(t, p)
	assert.Nil(t, m)
}

func BenchmarkMuxMatchStatic(b *testing.B) {
	mux := Mux{}
	mux.Handle("GET", "/foo/bar/baz", namedHandler{})
	mux.Handle("GET", "/foo/bar", namedHandler{})
	mux.Handle("GET", "/", namedHandler{})
	for i := 0; i < b.N; i++ {
		mux.match("GET", "/foo/bar")
	}
}

func BenchmarkMuxMatchStaticPrefix(b *testing.B) {
	mux := Mux{}
	mux.Handle("GET", "/foo/bar/baz", namedHandler{})
	mux.Handle("GET", "/foo/bar/", namedHandler{})
	mux.Handle("GET", "/", namedHandler{})
	for i := 0; i < b.N; i++ {
		mux.match("GET", "/foo/bar/rab")
	}
}

func TestMuxMatchVar(t *testing.T) {
	a := namedHandler{name: "a"}
	b := namedHandler{name: "b"}
	c := namedHandler{name: "c"}
	mux := Mux{}
	mux.Handle("GET", "/foo/:var/baz", a)
	mux.Handle("GET", "/:var/bar/", b)
	mux.Handle("GET", "/:var", c)
	m, p := mux.match("GET", "/foo/bar/baz")
	if assert.NotNil(t, m) {
		assert.Equal(t, Params{"var": "bar"}, p)
		assert.Equal(t, a, m.h)
	}
	m, p = mux.match("GET", "/foo/bar")
	if assert.NotNil(t, m) {
		assert.Equal(t, Params{"var": "foo"}, p)
		assert.Equal(t, b, m.h)
	}
	m, p = mux.match("GET", "/something")
	if assert.NotNil(t, m) {
		assert.Equal(t, Params{"var": "something"}, p)
		assert.Equal(t, c, m.h)
	}
	m, p = mux.match("GET", "/")
	assert.Nil(t, p)
	assert.Nil(t, m)
}

func BenchmarkMuxMatchVar(b *testing.B) {
	mux := Mux{}
	mux.Handle("GET", "/foo/:var/baz", namedHandler{})
	mux.Handle("GET", "/:var/bar/", namedHandler{})
	mux.Handle("GET", "/:var", namedHandler{})
	for i := 0; i < b.N; i++ {
		mux.match("GET", "/foo/bar")
	}
}
