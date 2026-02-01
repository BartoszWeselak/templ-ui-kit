package main

import (
	"context"
	"fmt"
	"net/http"

	"templ-ui-kit/view/layout"

	"github.com/a-h/templ"
)

func renderWithLayout(w http.ResponseWriter, content templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := layout.MainLayout(layout.LayoutProps{
		Content: content,
	}).Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Root route - przekierowanie na /admin
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	// Admin route
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		renderWithLayout(w, layout.AdminPage())
	})

	// Pricing route
	http.HandleFunc("/pricing", func(w http.ResponseWriter, r *http.Request) {
		renderWithLayout(w, layout.PricingPanel())
	})

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}
