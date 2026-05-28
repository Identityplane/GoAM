package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

func smtpAddress(host string, port int) string {
	if port <= 0 {
		port = 587
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func deliverSMTP(
	host string,
	port int,
	useImplicitTLS bool,
	username, password string,
	fromEmail string,
	envelopeTo []string,
	msg []byte,
) error {
	addr := smtpAddress(host, port)
	var auth smtp.Auth
	if strings.TrimSpace(username) != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	if useImplicitTLS {
		tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer c.Close()

		return sendWithClient(c, auth, fromEmail, envelopeTo, msg)
	}

	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	return sendWithClient(c, auth, fromEmail, envelopeTo, msg)
}

func sendWithClient(c *smtp.Client, auth smtp.Auth, fromEmail string, envelopeTo []string, msg []byte) error {
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
	}

	from := strings.TrimSpace(fromEmail)
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, rcpt := range envelopeTo {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt to %s: %w", rcpt, err)
		}
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	if err := c.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}
