package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/loganamcnichols/case_database/pkg/db"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/paymentintent"
	"github.com/stripe/stripe-go/v75/webhook"
)

type Item struct {
	Id     string `json:"id"`
	Amount string `json:"amount"`
}

func CalculateOrderAmount(items []Item) int64 {
	total := int64(0)
	for _, item := range items {
		if item.Id == "credits" {
			amt, err := strconv.Atoi(item.Amount)
			amt = int(math.Round(float64(amt) / 10))
			if err != nil {
				log.Printf("strconv.Atoi: %v", err)
				return 0
			}
			total += int64(amt)
		}
	}
	return total
}

func HandleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// bodyBytes, _ := io.ReadAll(r.Body)
	// log.Printf("bodyBytes: %v", string(bodyBytes))
	var req struct {
		Items []Item `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("Error getting session cookie")
	}
	if sessionID == nil {
		log.Printf("Error getting session cookie")
	}
	sessionMutex.RLock()
	var userID int64
	if sessionID != nil {
		userID = int64(sessionStore[sessionID.Value])
	}
	sessionMutex.RUnlock()

	// Create a PaymentIntent with amount and currency
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(CalculateOrderAmount(req.Items)),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		// In the latest version of the API, specifying the `automatic_payment_methods` parameter is optional because Stripe enables its functionality by default.
		Metadata: map[string]string{
			"user_id": fmt.Sprint(userID),
		},
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

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook error while parsing basic request. %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	endpointSecret := "whsec_754f4686510caeaeb4e04fe4028258d186bd12f4d77df6de130d8b0ec3087e4c"
	signatureHeader := r.Header.Get("Stripe-Signature")
	event, err = webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Successful payment for %d.", paymentIntent.Amount)
		con, err := db.Connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		userID, err := strconv.Atoi(paymentIntent.Metadata["user_id"])
		if err != nil {
			log.Printf("strconv.Atoi: %v", err)
			return
		}

		defer con.Close()
		err = db.UpdateUserCredits(con, userID, paymentIntent.Amount)
		if err != nil {
			log.Printf("Error updating user credits: %v", err)
		}

		// Then define and call a func to handle the successful payment intent.
		// handlePaymentIntentSucceeded(paymentIntent)
	// case "payment_method.attached":
	// 	var paymentMethod stripe.PaymentMethod
	// 	err := json.Unmarshal(event.Data.Raw, &paymentMethod)
	// 	if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// 	}
	// 	// Then define and call a func to handle the successful attachment of a PaymentMethod.
	// 	// handlePaymentMethodAttached(paymentMethod)
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}
