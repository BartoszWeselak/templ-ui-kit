package main

import (
	"fmt"
	"net/http"
	"templ-ui-kit/src/components"
	"templ-ui-kit/src/layouts"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content := components.Card(components.CardProps{
			Title:   "Witaj",
			Content: "Strona główna",
		})

		layouts.Base("Home", content).Render(r.Context(), w)
	})

	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		content := components.Card(components.CardProps{
			Title:   "Admin Panel",
			Content: "Tu będzie admin",
		})

		layouts.Base("Admin", content).Render(r.Context(), w)
	})

	fmt.Println("http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
