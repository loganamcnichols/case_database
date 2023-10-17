//go:build ignore
// +build ignore

package integrationtests

import (
	"os"
	"strings"
	"testing"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

func TestPurchaseAndDownload(t *testing.T) {
	os.Chdir("../")
	reqURL := "https://ecf.azd.uscourts.gov/doc1/025126869583"
	expectedSRC := "/cgi-bin/show_temp.pl?file=25125414"
	resDoc, err := scraper.PurchaseDocument(client, reqURL, "1348139", "9")
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
	_, err = scraper.PerformDownload(client, resDoc, reqURL, "1348139", "9")
	if err != nil {
		t.Fatalf("PerformDownload() returned error: %v", err)
	}
}

func TestGetDocumentSummary(t *testing.T) {
	requestURL := "https://ecf.almd.uscourts.gov/cgi-bin/DktRpt.pl?56135"
	respURL, err := scraper.GetFormURL(client, requestURL)
	if err != nil {
		t.Fatalf("GetDocumentURL() returned error: %v", err)
	}
	document, err := scraper.GetDocumentSummary(client, respURL, "56135")
	if err != nil {
		t.Fatalf("GetDocumentURL() returned error: %v", err)
	}
	headingElem := document.Find("h3").First()
	if !strings.Contains(headingElem.Text(), "2:14-cr-00646") {
		t.Fatalf("GetDocumentURL() returned incorrect document: %s", document.Find("#cmecfMainContent").Text())
	}
}
