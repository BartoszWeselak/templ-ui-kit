package main

import (
	"log"
	"net/http"

	"templ-ui-kit/src/layouts"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Render templ component
		ctx := r.Context()
		err := layouts.TestLayout(ctx).Render(ctx, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))))
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
