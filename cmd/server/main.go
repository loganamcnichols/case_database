package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/loganamcnichols/case_database/pkg/handlers"
	"github.com/stripe/stripe-go/v75"
)

func main() {
	stripe.Key = "sk_test_51NvmDbLe4GVYZHj7zS8EOZT8S1dOcXWVpsgHYRCWwTf2gaAM3PCytQsiUrq0Pr7EPT5q20DKNq6FipoUBIOScE5c00hobQBsAk"
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/pacer-lookup", handlers.PacerLookupHandler).Methods("GET") // Add this line
	r.HandleFunc("/checkout", handlers.CheckoutHandler).Methods("GET")
	r.HandleFunc("/buy-credits", handlers.BuyCreditsHandler).Methods("GET")
	r.HandleFunc("/credit-purchase-submit", handlers.BuyCreditsOnSubmit).Methods("POST")
	r.HandleFunc("/pacer-lookup-submit", handlers.PacerLookupOnSubmit).Methods("POST")
	r.HandleFunc("/pacer-lookup-case", handlers.PacerLookupCase).Methods("GET")
	r.HandleFunc("/pacer-lookup-docket-request", handlers.PacerLookupDocketRequest).Methods("POST")
	r.HandleFunc("/create-payment-intent", handlers.HandleCreatePaymentIntent)
	r.PathPrefix("/css/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "web/static"+r.URL.Path)
	})
	r.PathPrefix("/js/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "web/static"+r.URL.Path)
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
