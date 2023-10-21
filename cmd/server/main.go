package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/loganamcnichols/case_database/pkg/handlers"
	"github.com/stripe/stripe-go/v75"
)

func main() {
	stripe.Key = os.Getenv("STRIPE_SK")
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET", "POST")
	r.HandleFunc("/credits", handlers.CreditsHandler).Methods("GET")
	r.HandleFunc("/browse-docs", handlers.BrowseDocsHandler).Methods("GET")
	r.HandleFunc("/user-browse", handlers.UserBrowseHandler).Methods("GET")
	r.HandleFunc("/user-browse-search", handlers.UserBrowseSearchHandler).Methods("GET")
	r.HandleFunc("/user-browse-scroll", handlers.UserBrowseScrollHandler).Methods("GET")
	r.HandleFunc("/user-browse-docs", handlers.UserBrowseDocsHandler).Methods("GET")
	r.HandleFunc("/home", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/collapse-docs", handlers.DocsCollapseHandler).Methods("GET")
	r.HandleFunc("/purchase-doc", handlers.PurchaseDocHandler).Methods("POST")
	r.HandleFunc("/purchase-doc-credits", handlers.PurchaseDocCreditsHandler).Methods("GET")
	r.HandleFunc("/browse", handlers.BrowseHandler).Methods("GET")
	r.HandleFunc("/browse-search", handlers.BrowseSearchHandler).Methods("GET")
	r.HandleFunc("/browse-scroll", handlers.BrowseScrollHandler).Methods("GET")
	r.HandleFunc("/webhook", handlers.HandleWebhook).Methods("POST")
	r.HandleFunc("/pacer-lookup", handlers.PacerLookupHandler).Methods("GET") // Add this line
	r.HandleFunc("/checkout", handlers.CheckoutHandler).Methods("POST")
	r.HandleFunc("/buy-credits", handlers.BuyCreditsHandler).Methods("GET")
	r.HandleFunc("/credit-purchase-submit", handlers.BuyCreditsOnSubmit).Methods("POST")
	r.HandleFunc("/pacer-lookup-submit", handlers.PacerLookupOnSubmit).Methods("POST")
	r.HandleFunc("/pacer-lookup-case", handlers.PacerLookupCase).Methods("GET")
	r.HandleFunc("/pacer-login", handlers.PacerLoginHandler).Methods("GET", "POST")
	r.HandleFunc("/pacer-login-submit", handlers.PacerLoginSubmitHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/signup", handlers.SignupHandler).Methods("GET")
	r.HandleFunc("/pacer-lookup-summary-request", handlers.PacerLookupSummaryRequest).Methods("POST")
	r.HandleFunc("/view-docs", handlers.UserBrowseHandler).Methods("GET")
	r.HandleFunc("/signup-submit", handlers.SignupOnSubmitHandler).Methods("POST")
	r.HandleFunc("/login-submit", handlers.LoginOnSubmitHandler).Methods("POST")
	r.HandleFunc("/pacer-lookup-docket-request", handlers.PacerLookupDocketRequest).Methods("POST")
	r.HandleFunc("/create-payment-intent", handlers.HandleCreatePaymentIntent)
	r.PathPrefix("/css/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "web/static"+r.URL.Path)
	})
	r.PathPrefix("/img/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeFile(w, r, "web/static"+r.URL.Path)
	})
	r.PathPrefix("/pdfs/").HandlerFunc(handlers.ViewPDFHandler).Methods("GET")
	r.PathPrefix("/js/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "web/static"+r.URL.Path)
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
