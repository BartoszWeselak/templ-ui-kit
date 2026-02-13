package main

import (
	"log"
	"net/http"

	"templ-ui-kit/src/i18n"
	"templ-ui-kit/src/layouts"
)

func main() {
	// Initialize i18n
	_ = i18n.GetTranslator()

	// Setup routes
	mux := http.NewServeMux()

	// Main route with i18n middleware
	mux.Handle("/", i18n.LanguageMiddleware(http.HandlerFunc(handleIndex)))

	// Language change handler
	mux.HandleFunc("/set-language", handleSetLanguage)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))))

	log.Println("Server running on http://localhost:8080")
	log.Println("Available languages: en, pl, de, es, fr")
	log.Println("Change language: ?lang=pl or ?lang=en")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Render templ component with i18n support
	err := layouts.TestLayoutI18n(ctx).Render(ctx, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSetLanguage(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	// Set cookie
	i18n.SetLanguageCookie(w, lang)

	// Redirect back
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
