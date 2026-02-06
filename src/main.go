package main

import (
	"fmt"
	"net/http"
	"templ-ui-kit/src/components"
	"templ-ui-kit/src/layouts"
	"templ-ui-kit/src/pages"
)

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))))

	// Home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content := components.Card(components.CardProps{
			Title:   "Witaj w Templ UI Kit",
			Content: "Sprawdź dostępne strony w menu nawigacji",
		})

		layouts.BaseWithTheme("Home", content).Render(r.Context(), w)
	})

	// Button examples page
	http.HandleFunc("/examples/buttons", func(w http.ResponseWriter, r *http.Request) {
		layouts.BaseWithTheme("Button Examples", pages.ButtonExamples()).Render(r.Context(), w)
	})

	// Admin panel
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		stats := pages.AdminStats{
			Users:    150,
			Products: 342,
			Orders:   78,
			Revenue:  15234.50,
		}

		layouts.BaseWithTheme("Admin Panel", pages.AdminPanel(stats)).Render(r.Context(), w)
	})

	// Admin merchants page
	http.HandleFunc("/admin/merchants", func(w http.ResponseWriter, r *http.Request) {
		// Example data
		merchants := []components.MerchantListItem{
			{
				ID:        "1",
				Email:     "john@example.com",
				StoreName: "John's Store",
				StoreSlug: "johns-store",
				Status:    components.UserStatusActive,
			},
			{
				ID:        "2",
				Email:     "jane@example.com",
				StoreName: "Jane's Shop",
				StoreSlug: "janes-shop",
				Status:    components.UserStatusPending,
			},
			{
				ID:        "3",
				Email:     "bob@example.com",
				StoreName: "Bob's Market",
				StoreSlug: "bobs-market",
				Status:    components.UserStatusInactive,
			},
		}

		pageProps := pages.AdminMerchantsPageProps{
			InviteForm: components.AdminInviteMerchantFormProps{
				Data: components.AdminInviteMerchantFormData{
					Email:     "",
					StoreName: "",
					StoreSlug: "",
				},
				Errors: components.AdminInviteMerchantFormErrors{
					GenericMessage: "",
					Email:          "",
					StoreName:      "",
					StoreSlug:      "",
				},
				IsStateVisible: false,
			},
			MerchantList: components.MerchantListProps{
				Merchants:     merchants,
				CurrentPage:   1,
				TotalPages:    3,
				PaginationURL: "/admin/merchants?page=",
			},
			SuccessMessage: "",
			ErrorMessage:   "",
		}

		layouts.BaseWithTheme("Merchant Management", pages.AdminMerchantsPage(pageProps)).Render(r.Context(), w)
	})

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("- Home: http://localhost:8080")
	fmt.Println("- Button Examples: http://localhost:8080/examples/buttons")
	fmt.Println("- Admin: http://localhost:8080/admin")
	fmt.Println("- Merchants: http://localhost:8080/admin/merchants")
	http.ListenAndServe(":8080", nil)
}
