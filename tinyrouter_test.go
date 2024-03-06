package tinyrouter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {

	t.Run("route a simple GET request", func(t *testing.T) {
		router := NewRouter()

		req, err := http.NewRequest(http.MethodGet, "/home", nil)
		assertNoError(t, err)

		router.HandleFunc(http.MethodGet, "/home", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("home"))
		})

		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		assertEqual(t, res.Code, http.StatusOK)

		expected := "home"
		assertEqual(t, res.Body.String(), expected)

	})

	t.Run("not found", func(t *testing.T) {
		router := NewRouter()

		req, err := http.NewRequest("GET", "/notfound", nil)
		assertNoError(t, err)

		router.HandleFunc(http.MethodGet, "/found", func(_ http.ResponseWriter, _ *http.Request) {})

		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		assertEqual(t, res.Code, http.StatusNotFound)
		assertEqual(t, res.Body.String(), "404 page not found")
	})

	t.Run("get params from request", func(t *testing.T) {
		router := NewRouter()

		req, err := http.NewRequest(http.MethodGet, "/articles/4", nil)
		assertNoError(t, err)

		router.HandleFunc(http.MethodGet, "/articles/:id", func(w http.ResponseWriter, r *http.Request) {
			id := router.GetParam(r, "id")
			w.Write([]byte(id))
		})

		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		assertEqual(t, res.Code, http.StatusOK)

		expected := "4"
		assertEqual(t, res.Body.String(), expected)
	})

	t.Run("method not allowed", func(t *testing.T) {
		router := NewRouter()

		req, err := http.NewRequest(http.MethodGet, "/articles", nil)
		assertNoError(t, err)

		router.HandleFunc(http.MethodPost, "/articles", func(_ http.ResponseWriter, _ *http.Request) {})

		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		assertEqual(t, res.Code, http.StatusMethodNotAllowed)
		assertEqual(t, res.Header().Get("Allow"), "POST")
	})

	t.Run("redirect", func(t *testing.T) {
		router := NewRouter()
		RedirectTrailingSlash = true

		req, err := http.NewRequest(http.MethodGet, "/articles/", nil)
		assertNoError(t, err)

		router.HandleFunc(http.MethodGet, "/articles", func(_ http.ResponseWriter, _ *http.Request) {})

		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		assertEqual(t, res.Code, http.StatusSeeOther)
	})

}

func TestGroup(t *testing.T) {
	router := NewRouter()

	req, err := http.NewRequest(http.MethodPost, "/api/user/123", nil)
	assertNoError(t, err)

	router.Group("api/", func(tr *Router) {
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
