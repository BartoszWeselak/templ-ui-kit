package main

import (
	"fmt"
	"net/http"
	"templ-ui-kit/src/components"
	"templ-ui-kit/src/layouts"
	"templ-ui-kit/src/pages"
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
		stats := pages.AdminStats{
			Users:    150,
			Products: 342,
			Orders:   78,
			Revenue:  15234.50,
		}

		layouts.Base("Admin Panel", pages.AdminPanel(stats)).Render(r.Context(), w)
	})

	fmt.Println("http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
