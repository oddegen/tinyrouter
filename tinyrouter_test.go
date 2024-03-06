package tinyrouter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name       string
		method     string
		pattern    string
		path       string
		reqMethod  string
		handler    http.HandlerFunc
		statusCode int
		respBody   string
	}{
		{
			name:      "route a simple GET request",
			method:    http.MethodGet,
			pattern:   "/home",
			path:      "/home",
			reqMethod: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("home"))
			},
			statusCode: http.StatusOK,
			respBody:   "home",
		},
		{
			name:       "not found",
			method:     http.MethodGet,
			pattern:    "/found",
			path:       "/notfound",
			reqMethod:  http.MethodGet,
			handler:    func(_ http.ResponseWriter, _ *http.Request) {},
			statusCode: http.StatusNotFound,
			respBody:   "404 page not found\n",
		},
		{
			name:      "get named param from request",
			method:    http.MethodGet,
			pattern:   "/articles/:id",
			path:      "/articles/4",
			reqMethod: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				id := router.GetParam(r, "id")
				w.Write([]byte(id))
			},
			statusCode: http.StatusOK,
			respBody:   "4",
		},
		{
			name:      "get catch-all param from request",
			method:    http.MethodGet,
			pattern:   "/articles/*all",
			path:      "/articles/user/name/bob",
			reqMethod: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				all := router.GetParam(r, "all")
				w.Write([]byte(all))
			},
			statusCode: http.StatusOK,
			respBody:   "user/name/bob",
		},
		{
			name:      "empty param",
			method:    http.MethodGet,
			pattern:   "/articles/:id",
			path:      "/articles/123",
			reqMethod: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				name := router.GetParam(r, "name")
				w.Write([]byte(name))
			},
			statusCode: http.StatusOK,
			respBody:   "",
		},
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			pattern:    "/articles",
			path:       "/articles",
			reqMethod:  http.MethodPost,
			handler:    func(_ http.ResponseWriter, _ *http.Request) {},
			statusCode: http.StatusMethodNotAllowed,
			respBody:   "Method Not Allowed\n",
		},
		{
			name:       "no redirection",
			method:     http.MethodGet,
			pattern:    "/articles",
			path:       "/articles/",
			reqMethod:  http.MethodGet,
			handler:    func(_ http.ResponseWriter, _ *http.Request) {},
			statusCode: http.StatusNotFound,
			respBody:   "404 page not found\n",
		},
		{
			name:       "without preceding slash",
			method:     http.MethodGet,
			pattern:    "articles",
			path:       "/articles/",
			reqMethod:  http.MethodGet,
			handler:    func(_ http.ResponseWriter, _ *http.Request) {},
			statusCode: http.StatusOK,
			respBody:   "",
		},
	}

	for _, tt := range tests {
		router.HandleFunc(tt.method, tt.pattern, tt.handler)
	}

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := &http.Client{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest(tt.reqMethod, ts.URL+tt.path, nil)
			assertNoError(t, err)

			res, err := client.Do(req)
			assertNoError(t, err)

			assertEqual(t, res.StatusCode, tt.statusCode)
			body, err := io.ReadAll(res.Body)
			defer res.Body.Close()

			assertNoError(t, err)
			assertEqual(t, string(body), tt.respBody)
		})
	}

	t.Run("redirect", func(t *testing.T) {
		RedirectTrailingSlash = true

		router := NewRouter()
		router.HandleFunc(http.MethodGet, "/redirect", func(_ http.ResponseWriter, _ *http.Request) {})

		req, err := http.NewRequest(http.MethodGet, "/redirect/", nil)
		router := NewRouter()
		router.HandleFunc(http.MethodGet, "/redirect", func(_ http.ResponseWriter, _ *http.Request) {})

		req, err := http.NewRequest(http.MethodGet, "/redirect/", nil)
		assertNoError(t, err)

		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		assertEqual(t, res.Code, http.StatusSeeOther)
	})
}

func TestGroup(t *testing.T) {
	router := NewRouter()

	req, err := http.NewRequest(http.MethodPost, "/api/user/123", nil)
	assertNoError(t, err)

	router.Group("/api/", func(tr *Router) {
	router.Group("/api/", func(tr *Router) {
		tr.HandleFunc(http.MethodPost, "user/:id", func(w http.ResponseWriter, r *http.Request) {
			id := tr.GetParam(r, "id")
			w.Write([]byte(id))
		})
	})

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assertEqual(t, res.Code, http.StatusOK)

	expected := "123"
	assertEqual(t, res.Body.String(), expected)
}

func TestMiddleware(t *testing.T) {
	router := NewRouter()

	req, err := http.NewRequest(http.MethodGet, "/articles", nil)
	assertNoError(t, err)

	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("middleware"))
			next.ServeHTTP(w, r)
		})
	}

	writeMiddleware := Use(middlewareFunc)
	router.Handle(http.MethodGet, "/articles", writeMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})))

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	assertEqual(t, res.Code, http.StatusOK)

	expected := "middleware"
	assertEqual(t, res.Body.String(), expected)
}

func assertEqual[T comparable](t testing.TB, actual, expected T) {
	t.Helper()
	if actual != expected {
		t.Errorf("got %v, but want %v", actual, expected)
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
