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
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/webhook", handlers.HandleWebhook).Methods("POST")
	r.HandleFunc("/pacer-lookup", handlers.PacerLookupHandler).Methods("GET") // Add this line
	r.HandleFunc("/checkout", handlers.CheckoutHandler).Methods("POST")
	r.HandleFunc("/buy-credits", handlers.BuyCreditsHandler).Methods("GET")
	r.HandleFunc("/credit-purchase-submit", handlers.BuyCreditsOnSubmit).Methods("POST")
	r.HandleFunc("/pacer-lookup-submit", handlers.PacerLookupOnSubmit).Methods("POST")
	r.HandleFunc("/pacer-lookup-case", handlers.PacerLookupCase).Methods("GET")
	r.HandleFunc("/pacer-login", handlers.PacerLoginHandler).Methods("GET")
	r.HandleFunc("/pacer-login-submit", handlers.PacerLoginSubmitHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/signup", handlers.SignupHandler).Methods("GET")
	r.HandleFunc("/pacer-lookup-summary-request", handlers.PacerLookupSummaryRequest).Methods("POST")
	r.HandleFunc("/view-docs", handlers.ViewDocsHandler).Methods("GET")
	r.HandleFunc("/signup-submit", handlers.SignupOnSubmitHandler).Methods("POST")
	r.HandleFunc("/login-submit", handlers.LoginOnSubmitHandler).Methods("POST")
	r.HandleFunc("/pacer-lookup-docket-request", handlers.PacerLookupDocketRequest).Methods("POST")
	r.HandleFunc("/create-payment-intent", handlers.HandleCreatePaymentIntent)
	r.PathPrefix("/css/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "web/static"+r.URL.Path)
	})
	r.PathPrefix("/pdfs/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	r.PathPrefix("/js/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "web/static"+r.URL.Path)
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
