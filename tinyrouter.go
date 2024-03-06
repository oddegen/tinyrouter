package tinyrouter

import (
	"context"
	"net/http"
	"path"
	"strings"
)

var (
	RedirectTrailingSlash bool
)

type Router struct {
	prefix string
	routes []*route
}

type middleware func(http.Handler) http.Handler

func Use(middlewares ...middleware) middleware {
	return func(h http.Handler) http.Handler {
		for i := range middlewares {
			h = middlewares[len(middlewares)-1-i](h)
		}

		return h
	}
}

type route struct {
	method  string
	pattern string
	handler http.Handler
}

func NewRouter() *Router {
	return &Router{
		routes: make([]*route, 0),
	}
}

func (router *Router) Handle(method string, pattern string, handler http.Handler) {

	if method == "" {
		panic("tinyrouter: error no http method")
	}

	if handler == nil {
		panic("tinyrouter: error no handler")
	}

	validatePattern(pattern)
	pattern = cleanPath(pattern)

	if router.prefix != "" {
		pattern = path.Join(router.prefix, pattern)
	}

	route := &route{
		method:  method,
		pattern: pattern,
		handler: handler,
	}

	router.routes = append(router.routes, route)

}

func (router *Router) HandleFunc(method string, pattern string, handler http.HandlerFunc) {
	router.Handle(method, pattern, handler)
}

func (r *Router) Group(pattern string, fn func(r *Router)) {
	validatePattern(pattern)
	pattern = cleanPath(pattern)

	subrouter := NewRouter()
	subrouter.prefix = path.Join(r.prefix, pattern)

	fn(subrouter)

	r.routes = append(r.routes, subrouter.routes...)
}

type ctxKey string

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	allows := make([]string, 0)

	for _, route := range router.routes {
		path := cleanPath(r.URL.Path)

		if ok, trail := match(route.pattern, path); ok {
			if r.Method != route.method {
				allows = append(allows, route.method)
				continue
			}

			if trail && RedirectTrailingSlash {
				http.Redirect(w, r, removeTrailingSlash(path), http.StatusSeeOther)
				return
			} else if !trail {
				params := parseParams(route.pattern, path)
				ctx := context.WithValue(r.Context(), ctxKey("ctxKey"), params)
				route.handler.ServeHTTP(w, r.WithContext(ctx))
				return
			}
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

func match(pattern, path string) (match, trail bool) {
	key := make(map[string]struct{})

	if path[len(path)-1] == '/' {
		path = removeTrailingSlash(path)
		trail = true
	}

	for ; len(path) != 0 && len(pattern) != 0; pattern = pattern[1:] {
		switch pattern[0] {
		case '*', ':':
			param := getParam(pattern[1:])
			if _, ok := key[param]; ok {
				panic("tinyrouter: error wild-characters with identical names")
			}
			key[param] = struct{}{}

			if pattern[0] == '*' {
				path = path[len(path):]
			} else {
				path = path[len(getParam(path)):]
			}

			pattern = pattern[len(param):]
		case path[0]:
			path = path[1:]
		default:
			return false, trail
		}
	}
	return len(pattern) == 0 && len(path) == 0, trail
}

func parseParams(pattern, path string) (params map[string]string) {
	params = make(map[string]string)

	for ; len(path) != 0 && len(pattern) != 0; pattern = pattern[1:] {
		switch pattern[0] {
		case '*':
			k, v := getParam((pattern)[1:]), path
			params[k] = v
		case ':':
			k, v := getParam((pattern)[1:]), getParam(path)
			params[k] = v
		}
		path = path[1:]
	}

	return
}

func getParam(p string) string {
	index := strings.Index(p, "/")

	if index < 0 {
		index = len(p)
	}

	return p[:index]
}

func validatePattern(pattern string) {

	if len(pattern) < 1 {
		panic("tinyrouter: error no given pattern")
	}

	pattern = removeTrailingSlash(pattern)

	for i, c := range []byte(pattern) {
		if c != '*' && c != ':' {
			continue
		}

		if len(pattern[i+1:]) == 0 || pattern[i+1] == '/' {
			panic("tinyrouter: error un-named wildcard characters")
		}

		if c == '*' && strings.Contains(pattern[i:], "/") || c == '*' && strings.Contains(pattern[i:], ":") {
			panic("tinyrouter: error incorrect use of catch-all(*) wildcard matcher")
		}

		if c == ':' {
			if slash := strings.Index(pattern[i:], "/"); slash != -1 && strings.Contains(pattern[i+1:i+slash], "*") {
				panic("tinyrouter: error incorrect use of param(:) wildcard matcher")
			}
		}
	}

}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

func removeTrailingSlash(path string) string {
	if path[len(path)-1:] == "/" {
		path = path[:len(path)-1]
	}
	return path
}
