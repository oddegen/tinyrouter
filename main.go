package main

import (
	"log"
	"net/http"
	tinyrouter "oddegen/tinyrouter/pkg"
)

func main() {
	tinyrouter.RedirectTrailingSlash = true

	router := tinyrouter.NewRouter()

	router.HandleFunc(http.MethodGet, "/movie/:id/genre/action", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET :8000 /movie/:id GOT MOVIE WITH ID: " + router.GetParam(r, "id")))
	})

	router.HandleFunc(http.MethodPost, "/movie", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST :8000 INSERTED MOVIE"))
	})

	router.Group("/articles", func(tr *tinyrouter.Router) {

		tr.HandleFunc(http.MethodPost, "/*wild", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("POST :8000 INSERTED ARTICLE" + tr.GetParam(r, "wild")))
		})

		tr.HandleFunc("http.MethodGet", "/search", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("GET :8000 SEARCHING ARTICLES"))
		})

		tr.Group("/:articlesId", func(tr *tinyrouter.Router) {
			tr.HandleFunc(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("GET :8000 SEARCHING ARTICLES"))
			})
			tr.HandleFunc(http.MethodPut, "/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("PUT :8000 UPDATING ARTICLES"))
			})
			tr.HandleFunc(http.MethodDelete, "/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("DELETE :8000 DELETING ARTICLES"))
			})
		})
	})

	srv := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	log.Println("Listening on port: 8000")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
