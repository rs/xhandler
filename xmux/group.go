package xmux

import (
	"net/http"

	"github.com/rs/xhandler"
)

// Group makes it simple to configure a group of routes with the
// same prefix. Use mux.NewGroup("/prefix") to create a group.
type Group struct {
	m *Mux
	p string
}

func newRouteGroup(mux *Mux, path string) *Group {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	//Strip traling / (if present) as all added sub paths must start with a /
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return &Group{m: mux, p: path}
}

// NewGroup creates a new sub routes group with the provided path prefix.
// All routes added to the returned group will have the path prepended.
func (g *Group) NewGroup(path string) *Group {
	return newRouteGroup(g.m, g.subPath(path))
}

// GET is a shortcut for g.Handle("GET", path, handler)
func (g *Group) GET(path string, handler xhandler.HandlerC) {
	g.HandleC("GET", path, handler)
}

// HEAD is a shortcut for g.Handle("HEAD", path, handler)
func (g *Group) HEAD(path string, handler xhandler.HandlerC) {
	g.HandleC("HEAD", path, handler)
}

// OPTIONS is a shortcut for g.Handle("OPTIONS", path, handler)
func (g *Group) OPTIONS(path string, handler xhandler.HandlerC) {
	g.HandleC("OPTIONS", path, handler)
}

// POST is a shortcut for g.Handle("POST", path, handler)
func (g *Group) POST(path string, handler xhandler.HandlerC) {
	g.HandleC("POST", path, handler)
}

// PUT is a shortcut for g.Handle("PUT", path, handler)
func (g *Group) PUT(path string, handler xhandler.HandlerC) {
	g.HandleC("PUT", path, handler)
}

// PATCH is a shortcut for g.Handle("PATCH", path, handler)
func (g *Group) PATCH(path string, handler xhandler.HandlerC) {
	g.HandleC("PATCH", path, handler)
}

// DELETE is a shortcut for g.Handle("DELETE", path, handler)
func (g *Group) DELETE(path string, handler xhandler.HandlerC) {
	g.HandleC("DELETE", path, handler)
}

// HandleC registers a net/context aware request handler with the given
// path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (g *Group) HandleC(method, path string, handler xhandler.HandlerC) {
	g.m.HandleC(method, g.subPath(path), handler)
}

// Handle regiester a standard http.Handler request handler with the given
// path and method. With this adapter, your handler won't have access to the
// context and thus won't work with URL parameters.
func (g *Group) Handle(method, path string, handler http.Handler) {
	g.m.Handle(method, g.subPath(path), handler)
}

// HandleFunc regiester a standard http.HandlerFunc request handler with the given
// path and method. With this adapter, your handler won't have access to the
// context and thus won't work with URL parameters.
func (g *Group) HandleFunc(method, path string, handler http.HandlerFunc) {
	g.m.HandleFunc(method, g.subPath(path), handler)
}

func (g *Group) subPath(path string) string {
	if path[0] != '/' {
		panic("path must start with a '/'")
	}
	return g.p + path
}
