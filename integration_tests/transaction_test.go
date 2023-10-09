//go:build manualtest
// +build manualtest

package integrationtests

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

var clientTransaction *http.Client

func TestMainTransaction(m *testing.M) {
	// setup code
	var err error
	clientTransaction, err = scraper.LoginToPacer()
	if err != nil {
		fmt.Println("Error logging in to PACER")
		os.Exit(1)
	}
	m.Run()
}
func TestPurchaseAndDownload(t *testing.T) {
	reqURL := "https://ecf.azd.uscourts.gov/doc1/025126869583"
	expectedSRC := "/cgi-bin/show_temp.pl?file=25125414"
	resDoc, err := scraper.PurchaseDocument(clientTransaction, reqURL, "1348139", "9")
	if err != nil {
		t.Fatalf("PerformDownload() returned error: %v", err)
	}
	src, exists := resDoc.Find("iframe").First().Attr("src")
	if !exists {
		t.Fatalf("PerformDownload() returned incorrect document: %s", src)
	}
	if !strings.Contains(src, expectedSRC) {
		t.Fatalf("PerformDownload() returned incorrect document: %s", src)
	}
	err = scraper.PerformDownload(clientTransaction, resDoc, reqURL, "1348139", "9")
	if err != nil {
		t.Fatalf("PerformDownload() returned error: %v", err)
	}
}
