package email

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/Identityplane/GoAM/pkg/model"
)

func formatAddress(a model.EmailAddress) string {
	name := strings.TrimSpace(a.Name)
	emailAddr := strings.TrimSpace(a.Email)
	if emailAddr == "" {
		return ""
	}
	if name == "" {
		return emailAddr
	}
	return fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", name), emailAddr)
}

func collectRecipientEmails(addrs []model.EmailAddress) []string {
	var out []string
	for _, a := range addrs {
		e := strings.TrimSpace(a.Email)
		if e != "" {
			out = append(out, e)
		}
	}
	return out
}

func randomBoundary() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "goam-mime-boundary"
	}
	return "goam-" + hex.EncodeToString(b[:])
}

func writeQuotedPrintable(buf *bytes.Buffer, body string) error {
	w := quotedprintable.NewWriter(buf)
	if _, err := w.Write([]byte(body)); err != nil {
		return err
	}
	return w.Close()
}

func buildRFC5322Message(
	subject string,
	from model.EmailAddress,
	replyTo *model.EmailAddress,
	to, cc, bcc []model.EmailAddress,
	textBody, htmlBody string,
) ([]byte, error) {
	if strings.TrimSpace(from.Email) == "" {
		return nil, fmt.Errorf("from address is required")
	}
	if len(collectRecipientEmails(to)) == 0 {
		return nil, fmt.Errorf("at least one To recipient is required")
	}
	if strings.TrimSpace(textBody) == "" && strings.TrimSpace(htmlBody) == "" {
		return nil, fmt.Errorf("message must include text and/or html body")
	}

	var buf bytes.Buffer
	fromHeader := formatAddress(from)
	if _, err := mail.ParseAddress(fromHeader); err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	buf.WriteString("From: " + fromHeader + "\r\n")
	buf.WriteString("To: " + joinHeaderAddresses(to) + "\r\n")
	if len(cc) > 0 {
		buf.WriteString("Cc: " + joinHeaderAddresses(cc) + "\r\n")
	}
	if replyTo != nil && strings.TrimSpace(replyTo.Email) != "" {
		buf.WriteString("Reply-To: " + formatAddress(*replyTo) + "\r\n")
	}
	buf.WriteString("Subject: " + encodeSubjectLine(subject) + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")

	hasText := strings.TrimSpace(textBody) != ""
	hasHTML := strings.TrimSpace(htmlBody) != ""

	if hasText && hasHTML {
		alt := randomBoundary()
		buf.WriteString("Content-Type: multipart/alternative; boundary=" + alt + "\r\n\r\n")

		buf.WriteString("--" + alt + "\r\n")
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		if err := writeQuotedPrintable(&buf, textBody); err != nil {
			return nil, err
		}
		buf.WriteString("\r\n")

		buf.WriteString("--" + alt + "\r\n")
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		if err := writeQuotedPrintable(&buf, htmlBody); err != nil {
			return nil, err
		}
		buf.WriteString("\r\n--" + alt + "--\r\n")
	} else if hasHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		if err := writeQuotedPrintable(&buf, htmlBody); err != nil {
			return nil, err
		}
		buf.WriteString("\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		if err := writeQuotedPrintable(&buf, textBody); err != nil {
			return nil, err
		}
		buf.WriteString("\r\n")
	}

	return buf.Bytes(), nil
}

func joinHeaderAddresses(addrs []model.EmailAddress) string {
	parts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		if s := formatAddress(a); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ", ")
}

func envelopeFromEmail(from model.EmailAddress) string {
	return strings.TrimSpace(from.Email)
}

func encodeSubjectLine(s string) string {
	for _, r := range s {
		if r > 127 || r == '\n' || r == '\r' {
			return mime.QEncoding.Encode("UTF-8", s)
		}
	}
	return s
}

func allEnvelopeRecipients(to, cc, bcc []model.EmailAddress) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, list := range [][]model.EmailAddress{to, cc, bcc} {
		for _, e := range collectRecipientEmails(list) {
			if _, ok := seen[e]; ok {
				continue
			}
			seen[e] = struct{}{}
			out = append(out, e)
		}
	}
	return out
}
