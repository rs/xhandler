package xhandler

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

type Mux struct {
	m []muxEntry
}

// Params holds parameters extracted from the route
type Params map[string]string

type muxEntry struct {
	m string
	h HandlerC
	p patternDef
}

type pos struct {
	start uint16
	end   uint16
}

type patternDef struct {
	content    string
	len        int
	components []pos
	vars       []bool
	isPrefix   bool
	isStatic   bool
}

type key int

const paramKey key = iota

func newParamContext(ctx context.Context, p Params) context.Context {
	return context.WithValue(ctx, paramKey, p)
}

// URLParams returns URL parameters stored in context
func URLParams(ctx context.Context) Params {
	if ctx == nil {
		return Params{}
	}
	if p, ok := ctx.Value(paramKey).(Params); ok {
		return p
	}
	return Params{}
}

// NewMux returns a new muxer instance
func NewMux() *Mux {
	return &Mux{}
}

// Handle registers the handler for the given pattern.
func (mux *Mux) Handle(method, pattern string, handler HandlerC) {
	mux.m = append(mux.m, muxEntry{
		m: strings.ToUpper(method),
		p: newPattern(pattern),
		h: handler,
	})
}

// ServeHTTPC implements xhandler.HandlerC interface
func (mux *Mux) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if m, p := mux.match(r.Method, r.URL.Path); m != nil {
		if p != nil && len(p) > 0 {
			ctx = newParamContext(ctx, p)
			m.h.ServeHTTPC(ctx, w, r)
		}
	}
}

func (mux *Mux) match(method, path string) (m *muxEntry, p Params) {
	comps := parsePath(path)
	len := -1
	// Find the longest (most components) matching path
	for i, _m := range mux.m {
		if _m.m == method && matchPattern(path, comps, _m.p) && _m.p.len > len {
			len = _m.p.len
			m = &mux.m[i]
		}
	}
	// Gather variables if any
	if m != nil && !m.p.isStatic {
		p = Params{}
		for i, v := range m.p.vars {
			if v {
				// Component position in pattern for the variable name
				comp := m.p.components[i]
				// Var name is the component's content without the : prefix
				name := m.p.content[comp.start+1 : comp.end]
				// Component position in path for the variable's content
				comp = comps[i]
				val := path[comp.start:comp.end]
				p[name] = val
			}
		}
	}
	return
}

func matchPattern(path string, components []pos, pat patternDef) bool {
	if len(components) < pat.len {
		return false
	}
	for i, p := range components {
		if i > pat.len-1 {
			return pat.isPrefix
		}
		if pat.vars[i] {
			continue
		}
		pp := pat.components[i]
		if path[p.start:p.end] != pat.content[pp.start:pp.end] {
			return false
		}
	}
	return true
}

// parsePath finds all the ranges for path components
func parsePath(path string) []pos {
	p := make([]pos, 0, 5)
	for i, j, l := 0, 0, len(path); i <= l; i++ {
		j = strings.IndexByte(path[i:], '/')
		if j == -1 {
			if l-i > 0 {
				p = append(p, pos{start: uint16(i), end: uint16(l)})
			}
			break
		} else if j > 0 {
			p = append(p, pos{start: uint16(i), end: uint16(i + j)})
		}
		i += j
	}
	return p
}

func newPattern(pattern string) patternDef {
	c := parsePath(pattern)
	l := len(c)
	pat := patternDef{
		content:    pattern,
		components: c,
		vars:       make([]bool, l, l),
		len:        l,
		// A pattern is a prefix if it ends with a / (and is not just /)
		isPrefix: pattern != "/" && pattern[len(pattern)-1] == '/',
		isStatic: true,
	}
	for i, p := range c {
		if pattern[p.start] == ':' && p.end-p.start > 1 {
			pat.vars[i] = true
			pat.isStatic = false
		}
	}
	return pat
}
