package integration

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// extractStepFromHTML extracts the step value from the hidden input field in the HTML response
func extractStepFromHTML(t *testing.T, htmlContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	stepInput := doc.Find("input[type='hidden'][name='step']")
	if stepInput.Length() == 0 {
		t.Fatal("Expected to find hidden step field in HTML response")
	}

	stepValue, exists := stepInput.Attr("value")
	if !exists {
		t.Fatal("Step input field found but has no value attribute")
	}

	return stepValue
}

