package main

import (
	"log"
	"net/http"
	tinyrouter "oddegen/tinyrouter/pkg"
)

func main() {
	router := tinyrouter.NewRouter()

	router.Handle(http.MethodGet, "/movie/:id/genre/action", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET :8000 /movie/:id GOT MOVIE WITH ID: " + router.GetParam(r, "id")))
	})

	router.Handle(http.MethodPost, "/movie", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST :8000 INSERTED MOVIE"))
	})

	router.Route("/articles", func(tr *tinyrouter.Router) {

		tr.Handle(http.MethodPost, "/*wild", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("POST :8000 INSERTED ARTICLE" + tr.GetParam(r, "wild")))
		})

		tr.Handle("http.MethodGet", "/search", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("GET :8000 SEARCHING ARTICLES"))
		})

		tr.Route("/:articlesId", func(tr *tinyrouter.Router) {
			tr.Handle(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("GET :8000 SEARCHING ARTICLES"))
			})
			tr.Handle(http.MethodPut, "/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("PUT :8000 UPDATING ARTICLES"))
			})
			tr.Handle(http.MethodDelete, "/", func(w http.ResponseWriter, r *http.Request) {
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
