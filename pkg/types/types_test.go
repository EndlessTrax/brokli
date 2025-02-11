package types

import (
	"testing"
)

var testTag = AnchorTag{
	RawLink: `<a href="https://www.google.com">Google</a>`,
	Href:    "https://www.google.com",
	Text:    "Google",
	Status:  200,
}


func main(t *testing.T) {
	testAnchorTag(t)
}

func testAnchorTag(t *testing.T) {
	if testTag.RawLink != `<a href="https://www.google.com">Google</a>` {
		t.Errorf("RawLink does not match")
	}
	if testTag.Href != "https://www.google.com" {
		t.Errorf("Href does not match")
	}
	if testTag.Text != "Google" {
		t.Errorf("Text does not match")
	}
	if testTag.Status != 200 {
		t.Errorf("Status does not match")
	}
}