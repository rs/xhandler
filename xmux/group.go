package xmux

import "github.com/rs/xhandler"

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
	g.Handle("GET", path, handler)
}

// HEAD is a shortcut for g.Handle("HEAD", path, handler)
func (g *Group) HEAD(path string, handler xhandler.HandlerC) {
	g.Handle("HEAD", path, handler)
}

// OPTIONS is a shortcut for g.Handle("OPTIONS", path, handler)
func (g *Group) OPTIONS(path string, handler xhandler.HandlerC) {
	g.Handle("OPTIONS", path, handler)
}

// POST is a shortcut for g.Handle("POST", path, handler)
func (g *Group) POST(path string, handler xhandler.HandlerC) {
	g.Handle("POST", path, handler)
}

// PUT is a shortcut for g.Handle("PUT", path, handler)
func (g *Group) PUT(path string, handler xhandler.HandlerC) {
	g.Handle("PUT", path, handler)
}

// PATCH is a shortcut for g.Handle("PATCH", path, handler)
func (g *Group) PATCH(path string, handler xhandler.HandlerC) {
	g.Handle("PATCH", path, handler)
}

// DELETE is a shortcut for g.Handle("DELETE", path, handler)
func (g *Group) DELETE(path string, handler xhandler.HandlerC) {
	g.Handle("DELETE", path, handler)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (g *Group) Handle(method, path string, handler xhandler.HandlerC) {
	g.m.Handle(method, g.subPath(path), handler)
}

func (g *Group) subPath(path string) string {
	if path[0] != '/' {
		panic("path must start with a '/'")
	}
	return g.p + path
}
