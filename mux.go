package xhandler

import (
	"net/http"

	"golang.org/x/net/context"
)

// Mux is a xhandler.HandlerC which can be used to dispatch requests to different
// handler functions via configurable routes
//
// The muxer matches incoming requests by the request method and the path.
// If a handle is registered for this path and method, the router delegates the
// request to that function.
// For the methods GET, POST, PUT, PATCH and DELETE shortcut functions exist to
// register handlers, for all other methods mux.Handle can be used.
//
// The registered path, against which the muxer matches incoming requests, can
// contain two types of parameters:
//  Syntax    Type
//  :name     named parameter
//  *name     catch-all parameter
//
// Named parameters are dynamic path segments. They match anything until the
// next '/' or the path end:
//  Path: /blog/:category/:post
//
//  Requests:
//   /blog/go/request-routers            match: category="go", post="request-routers"
//   /blog/go/request-routers/           no match, but the router would redirect
//   /blog/go/                           no match
//   /blog/go/request-routers/comments   no match
//
// Catch-all parameters match anything until the path end, including the
// directory index (the '/' before the catch-all). Since they match anything
// until the end, catch-all parameters must always be the final path element.
//  Path: /files/*filepath
//
//  Requests:
//   /files/                             match: filepath="/"
//   /files/LICENSE                      match: filepath="/LICENSE"
//   /files/templates/article.html       match: filepath="/templates/article.html"
//   /files                              no match, but the router would redirect
//
// The value of parameters is saved as aParams type saved into the context.
// Parameters can be retrieved by name using xhandler.URLParams(ctx).Get(name) method:
//  user := xhandler.URLParams(ctx).Get("user") // defined by :user or *user
type Mux struct {
	trees map[string]*node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.Error with http.StatusNotFound is used.
	NotFound HandlerC

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	MethodNotAllowed HandlerC

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(context.Context, http.ResponseWriter, *http.Request, interface{})
}

// Params holds URL parameters.
type Params struct {
	params []struct {
		key   string
		value string
	}
}

// Get returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) Get(name string) string {
	for i := range ps.params {
		if ps.params[i].key == name {
			return ps.params[i].value
		}
	}
	return ""
}

type key int

const paramKey key = iota

var emptyParams = Params{}

func newParamContext(ctx context.Context, p Params) context.Context {
	return context.WithValue(ctx, paramKey, p)
}

// URLParams returns URL parameters stored in context
func URLParams(ctx context.Context) Params {
	if ctx == nil {
		return emptyParams
	}
	if p, ok := ctx.Value(paramKey).(Params); ok {
		return p
	}
	return emptyParams
}

// NewMux returns a new muxer instance
func NewMux() *Mux {
	return &Mux{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
	}
}

// GET is a shortcut for mux.Handle("GET", path, handler)
func (mux *Mux) GET(path string, handler HandlerC) {
	mux.Handle("GET", path, handler)
}

// HEAD is a shortcut for mux.Handle("HEAD", path, handler)
func (mux *Mux) HEAD(path string, handler HandlerC) {
	mux.Handle("HEAD", path, handler)
}

// OPTIONS is a shortcut for mux.Handle("OPTIONS", path, handler)
func (mux *Mux) OPTIONS(path string, handler HandlerC) {
	mux.Handle("OPTIONS", path, handler)
}

// POST is a shortcut for mux.Handle("POST", path, handler)
func (mux *Mux) POST(path string, handler HandlerC) {
	mux.Handle("POST", path, handler)
}

// PUT is a shortcut for mux.Handle("PUT", path, handler)
func (mux *Mux) PUT(path string, handler HandlerC) {
	mux.Handle("PUT", path, handler)
}

// PATCH is a shortcut for mux.Handle("PATCH", path, handler)
func (mux *Mux) PATCH(path string, handler HandlerC) {
	mux.Handle("PATCH", path, handler)
}

// DELETE is a shortcut for mux.Handle("DELETE", path, handler)
func (mux *Mux) DELETE(path string, handler HandlerC) {
	mux.Handle("DELETE", path, handler)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (mux *Mux) Handle(method, path string, handler HandlerC) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if mux.trees == nil {
		mux.trees = make(map[string]*node)
	}

	root := mux.trees[method]
	if root == nil {
		root = new(node)
		mux.trees[method] = root
	}

	root.addRoute(path, handler)
}

func (mux *Mux) recv(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if rcv := recover(); rcv != nil {
		mux.PanicHandler(ctx, w, r, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (mux *Mux) Lookup(method, path string) (HandlerC, Params, bool) {
	if root := mux.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, emptyParams, false
}

// ServeHTTPC implements xhandler.HandlerC interface
func (mux *Mux) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if mux.PanicHandler != nil {
		defer mux.recv(ctx, w, r)
	}

	if root := mux.trees[r.Method]; root != nil {
		path := r.URL.Path

		if handler, p, tsr := root.getValue(path); handler != nil {
			if len(p.params) > 0 {
				ctx = newParamContext(ctx, p)
			}
			handler.ServeHTTPC(ctx, w, r)
			return
		} else if r.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if r.Method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && mux.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					r.URL.Path = path[:len(path)-1]
				} else {
					r.URL.Path = path + "/"
				}
				http.Redirect(w, r, r.URL.String(), code)
				return
			}

			// Try to fix the request path
			if mux.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					mux.RedirectTrailingSlash,
				)
				if found {
					r.URL.Path = string(fixedPath)
					http.Redirect(w, r, r.URL.String(), code)
					return
				}
			}
		}
	}

	// Handle 405
	if mux.HandleMethodNotAllowed {
		for method := range mux.trees {
			// Skip the requested method - we already tried this one
			if method == r.Method {
				continue
			}

			handle, _, _ := mux.trees[method].getValue(r.URL.Path)
			if handle != nil {
				if mux.MethodNotAllowed != nil {
					mux.MethodNotAllowed.ServeHTTPC(ctx, w, r)
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if mux.NotFound != nil {
		mux.NotFound.ServeHTTPC(ctx, w, r)
	} else {
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
}
