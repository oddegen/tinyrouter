package tinyrouter

import (
	"context"
	"net/http"
	"strings"
)

// pattern: /movies/:id => /movies/12 || /movies/one
// pattern: /movies/* => /movies/two || /movies/category/action

type Router struct {
	routes map[string]route
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]route),
	}
}

type ctxKey string

type route struct {
	handler http.HandlerFunc
	method  string
	pattern string
}

func (r *Router) Handle(method, pattern string, handler http.HandlerFunc) {
	if len(pattern) > 1 && pattern[len(pattern)-1] == '/' {
		panic("tinyrouter: Error can't have a trailing /")
	}

	if pattern[0] != '/' {
		pattern = "/" + pattern
	}

	if len(pattern) == 1 && pattern[len(pattern)-1] == '/' {
		pattern = ""
	}

	route := route{
		method:  strings.ToUpper(method),
		pattern: pattern,
		handler: handler,
	}

	r.routes[pattern+method] = route
}

var params = make(map[string]string)

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allows := make([]string, 0)

	for _, route := range router.routes {

		if match(route.pattern, strings.TrimSuffix(r.URL.Path, "/")) {
			if r.Method != route.method {
				allows = append(allows, route.method)
				continue
			}
			ctx := context.WithValue(r.Context(), ctxKey("ctxKey"), params)
			route.handler(w, r.WithContext(ctx))
			return
		}

	}

	if len(allows) > 0 {
		w.Header().Set("Allow", strings.Join(allows, ", "))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, r)

}

func (route *Router) GetParam(r *http.Request, key string) string {
	field := r.Context().Value(ctxKey("ctxKey")).(map[string]string)
	return field[key]
}

func (r *Router) Route(pattern string, fn func(r *Router)) {
	subrouter := NewRouter()

	fn(subrouter)
	for _, route := range subrouter.routes {
		newPath := pattern + route.pattern
		r.Handle(route.method, newPath, route.handler)
	}
}

func match(pattern, path string) bool {

	for i := 1; len(path) != 0 && len(pattern) != 0; pattern = pattern[i:] {
		switch pattern[0] {
		case '*':
			if len(pattern) < 2 {
				panic("tinyrouter: Error no identifier for wildcard param")
			}

			i = len(pattern)
			params[pattern[1:]] = path
			path = path[len(path):]
		case ':':
			i = getParam(pattern)

			if len(pattern[:i]) < 2 {
				panic("tinyrouter: Error no identifier for named param")
			}

			if pattern[1] == '*' {
				panic("tinyrouter: Error can't use wildcard (*) after a named param")
			}

			if len(path) == 0 {
				continue
			}

			params[pattern[1:i]] = path[:getParam(path)]

			path = path[getParam(path):]
		case path[0]:
			i = 1
			path = path[1:]
		default:
			return false
		}
	}
	return len(pattern) == 0 && len(path) == 0
}

func getParam(pattern string) int {
	i := strings.IndexByte(pattern, '/')

	if i < 0 {
		i = len(pattern)
	}

	return i
}
