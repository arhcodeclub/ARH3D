package auth

import (
	"fmt"
    "log"
    "net/smtp"
	"os"
)

func SendMagicLinkEmail(toEmail, link string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if host == "" || port == "" || user == "" || pass == "" || from == "" {
        fmt.Printf("[EMAIL] Magic login link for %s: %s.\n", toEmail, link) // TODO: remove in prod.
		return nil
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	subject := "Your ARH3D login link"
	body := fmt.Sprintf("Click the link below to log in:\n\n%s\n\nThis link expires in 10 minutes.", link)

	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

    log.Printf("[EMAIL] Sending magic link to %s via %s.", toEmail, addr)
	if err := smtp.SendMail(addr, auth, from, []string{toEmail}, msg); err != nil {
		log.Printf("[EMAIL] Error sending magic link to %s: %v", toEmail, err)
		return err
	}

	log.Printf("[EMAIL] Successfully sent magic link to %s.", toEmail)
	return nil
}
