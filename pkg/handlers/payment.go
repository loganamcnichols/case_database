package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/paymentintent"
)

type item struct {
	id string
}

func calculateOrderAmount(items []item) int64 {
	total := int64(0)
	for _, item := range items {
		if item.id == "1" {
			total += 100
		}
	}
	return total
}

func HandleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Items []item `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	// Create a PaymentIntent with amount and currency
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(calculateOrderAmount(req.Items)),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		// In the latest version of the API, specifying the `automatic_payment_methods` parameter is optional because Stripe enables its functionality by default.
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	log.Printf("pi.New: %v", pi.ClientSecret)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("pi.New: %v", err)
		return
	}

	writeJSON(w, struct {
		ClientSecret string `json:"clientSecret"`
	}{
		ClientSecret: pi.ClientSecret,
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}
