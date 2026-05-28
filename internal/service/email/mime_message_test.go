package email

import (
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
)

func TestFormatAddress(t *testing.T) {
	if got := formatAddress(model.EmailAddress{Email: "a@b.co"}); got != "a@b.co" {
		t.Fatalf("email only: got %q", got)
	}
	got := formatAddress(model.EmailAddress{Email: "a@b.co", Name: "Ann Example"})
	if !strings.Contains(got, "a@b.co") || !strings.Contains(got, "Ann") {
		t.Fatalf("with name: got %q", got)
	}
}

func TestBuildRFC5322Message_multipart(t *testing.T) {
	from := model.EmailAddress{Email: "from@x.test", Name: "Sender"}
	to := []model.EmailAddress{{Email: "to@x.test", Name: "To User"}}
	msg, err := buildRFC5322Message("Hello subject", from, nil, to, nil, nil, "plain body", "<p>hi</p>")
	if err != nil {
		t.Fatal(err)
	}
	raw := string(msg)
	if !strings.Contains(raw, "multipart/alternative") {
		t.Fatalf("expected multipart: %s", raw)
	}
	if !strings.Contains(raw, "plain body") || !strings.Contains(raw, "<p>hi</p>") {
		t.Fatalf("missing bodies: %s", raw)
	}
}

func TestBuildRFC5322Message_requiresTo(t *testing.T) {
	from := model.EmailAddress{Email: "from@x.test"}
	_, err := buildRFC5322Message("s", from, nil, nil, nil, nil, "t", "")
	if err == nil {
		t.Fatal("expected error")
	}
}
