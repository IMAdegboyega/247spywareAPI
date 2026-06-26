package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// SMTPSender is a thin wrapper over net/smtp for one provider's credentials.
// We keep two of these inside EmailService — one for transactional mail
// (Gmail) and one for newsletter blasts (Resend).
type SMTPSender struct {
	label    string // "notification" or "newsletter" — only for log lines
	host     string
	port     string
	username string
	password string
	from     string
	fromName string
	enabled  bool
}

func newSMTPSender(label, hostEnv, portEnv, userEnv, passEnv, fromEnv, fromNameEnv string) *SMTPSender {
	host := os.Getenv(hostEnv)
	port := os.Getenv(portEnv)
	username := os.Getenv(userEnv)
	password := os.Getenv(passEnv)
	from := os.Getenv(fromEnv)
	fromName := os.Getenv(fromNameEnv)

	if from == "" {
		from = username
	}
	if port == "" {
		port = "587"
	}
	if fromName == "" {
		fromName = "247 Tech Spyware"
	}

	enabled := host != "" && username != "" && password != ""

	if enabled {
		log.Printf("Email[%s] enabled: %s:%s from %q <%s>", label, host, port, fromName, from)
	} else {
		log.Printf("Email[%s] disabled: %s/%s/%s missing", label, hostEnv, userEnv, passEnv)
	}

	return &SMTPSender{
		label:    label,
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		fromName: fromName,
		enabled:  enabled,
	}
}

// Send delivers an HTML email. Returns nil silently if the sender is not
// configured — callers shouldn't error on a missing SMTP setup.
func (s *SMTPSender) Send(to, subject, htmlBody string) error {
	if !s.enabled {
		return nil
	}

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	headers := []string{
		fmt.Sprintf("From: %s <%s>", s.fromName, s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	msg := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody)

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, msg); err != nil {
		log.Printf("Email[%s] failed -> %s: %v", s.label, to, err)
		return err
	}

	log.Printf("Email[%s] sent -> %s: %s", s.label, to, subject)
	return nil
}

// EmailService bundles both senders behind one interface so callers don't have
// to know which provider is used for which message type.
type EmailService struct {
	notification *SMTPSender // Gmail: password resets, admin/author notifications
	newsletter   *SMTPSender // Resend: subscriber blasts
}

func NewEmailService() *EmailService {
	return &EmailService{
		notification: newSMTPSender(
			"notification",
			"NOTIFICATION_SMTP_HOST",
			"NOTIFICATION_SMTP_PORT",
			"NOTIFICATION_SMTP_USERNAME",
			"NOTIFICATION_SMTP_PASSWORD",
			"NOTIFICATION_SMTP_FROM",
			"NOTIFICATION_SMTP_FROM_NAME",
		),
		newsletter: newSMTPSender(
			"newsletter",
			"NEWSLETTER_SMTP_HOST",
			"NEWSLETTER_SMTP_PORT",
			"NEWSLETTER_SMTP_USERNAME",
			"NEWSLETTER_SMTP_PASSWORD",
			"NEWSLETTER_SMTP_FROM",
			"NEWSLETTER_SMTP_FROM_NAME",
		),
	}
}

// IsEnabled reports whether *either* sender is configured. The newsletter
// path uses its own IsNewsletterEnabled / IsNotificationEnabled below.
func (s *EmailService) IsEnabled() bool {
	return s.notification.enabled || s.newsletter.enabled
}

func (s *EmailService) IsNotificationEnabled() bool { return s.notification.enabled }
func (s *EmailService) IsNewsletterEnabled() bool   { return s.newsletter.enabled }

// siteURL / apiURL pulled together so every template renders the same hosts.
func siteURL() string {
	v := os.Getenv("SITE_URL")
	if v == "" {
		v = "http://localhost:3000"
	}
	return strings.TrimRight(v, "/")
}

func apiURL() string {
	v := os.Getenv("API_URL")
	if v == "" {
		v = "http://localhost:8080"
	}
	return strings.TrimRight(v, "/")
}

// ===========================================================================
// Newsletter (Resend)
// ===========================================================================

// SendWelcomeEmail greets a newly subscribed reader and includes an unsubscribe
// link so they can back out without ever needing to email us.
func (s *EmailService) SendWelcomeEmail(toEmail, unsubscribeToken string) error {
	if !s.newsletter.enabled {
		return nil
	}

	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe?token=%s", apiURL(), unsubscribeToken)
	subject := "Welcome to 24|7 Tech Spyware"

	body := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; padding: 20px 0; border-bottom: 2px solid #FF9C00;">
    <h1 style="color: #231F1F; margin: 0;">247 Tech Spyware</h1>
  </div>
  <div style="padding: 30px 0;">
    <h2 style="color: #231F1F;">You're in.</h2>
    <p style="color: #555; line-height: 1.6;">Thanks for subscribing. We'll send you a heads-up whenever a new article goes live — straight tech news, no fluff, no spam.</p>
    <p style="color: #555; line-height: 1.6;">If this wasn't you, or you change your mind, you can leave any time using the unsubscribe link at the bottom of every email.</p>
    <div style="text-align: center; padding: 20px 0;">
      <a href="%s" style="background-color: #FF9C00; color: #231F1F; padding: 12px 30px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">Visit the site</a>
    </div>
  </div>
  <div style="border-top: 1px solid #eee; padding-top: 15px; text-align: center;">
    <p style="color: #999; font-size: 12px;">You're receiving this because you just subscribed to our newsletter.<br>
      <a href="%s" style="color: #FF9C00;">Unsubscribe</a>
    </p>
  </div>
</body></html>`, siteURL(), unsubscribeURL)

	return s.newsletter.Send(toEmail, subject, body)
}

func (s *EmailService) SendNewPostNotification(toEmail, postTitle, postSlug, unsubscribeToken string) error {
	if !s.newsletter.enabled {
		return nil
	}

	postURL := fmt.Sprintf("%s/blog/%s", siteURL(), postSlug)
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe?token=%s", apiURL(), unsubscribeToken)

	subject := fmt.Sprintf("New Post: %s", postTitle)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; padding: 20px 0; border-bottom: 2px solid #FF9C00;">
    <h1 style="color: #231F1F; margin: 0;">247 Tech Spyware</h1>
  </div>
  <div style="padding: 30px 0;">
    <h2 style="color: #231F1F;">%s</h2>
    <p style="color: #555; line-height: 1.6;">A new article has been published on 247 Tech Spyware. Check it out!</p>
    <div style="text-align: center; padding: 20px 0;">
      <a href="%s" style="background-color: #FF9C00; color: #231F1F; padding: 12px 30px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">Read Article</a>
    </div>
  </div>
  <div style="border-top: 1px solid #eee; padding-top: 15px; text-align: center;">
    <p style="color: #999; font-size: 12px;">You received this because you subscribed to our newsletter.<br>
      <a href="%s" style="color: #FF9C00;">Unsubscribe</a>
    </p>
  </div>
</body></html>`, postTitle, postURL, unsubscribeURL)

	return s.newsletter.Send(toEmail, subject, body)
}

// ===========================================================================
// Notification emails (Gmail)
// ===========================================================================

// SendPasswordReset emails the reset link to whoever asked for it.
func (s *EmailService) SendPasswordReset(toEmail, displayName, token string) error {
	if !s.notification.enabled {
		return nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", siteURL(), token)
	subject := "Reset your 247 Tech Spyware password"

	body := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; padding: 20px 0; border-bottom: 2px solid #FF9C00;">
    <h1 style="color: #231F1F; margin: 0;">247 Tech Spyware</h1>
  </div>
  <div style="padding: 30px 0;">
    <p style="color: #231F1F; font-size: 16px;">Hi %s,</p>
    <p style="color: #555; line-height: 1.6;">We received a request to reset the password on your account. Click the button below to choose a new password. This link expires in 1 hour.</p>
    <div style="text-align: center; padding: 20px 0;">
      <a href="%s" style="background-color: #FF9C00; color: #231F1F; padding: 12px 30px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">Reset Password</a>
    </div>
    <p style="color: #777; font-size: 13px; line-height: 1.6;">If the button doesn't work, copy and paste this URL into your browser:<br><span style="word-break: break-all;">%s</span></p>
    <p style="color: #777; font-size: 13px; line-height: 1.6;">If you didn't ask for this, you can safely ignore the email — your password won't change.</p>
  </div>
</body></html>`, displayName, resetURL, resetURL)

	return s.notification.Send(toEmail, subject, body)
}

// SendAdminAlert is the generic "FYI, something happened" mail to the admin.
func (s *EmailService) SendAdminAlert(toEmail, subject, headline, body string) error {
	if !s.notification.enabled {
		return nil
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; padding: 20px 0; border-bottom: 2px solid #FF9C00;">
    <h1 style="color: #231F1F; margin: 0;">247 Tech Spyware</h1>
    <p style="color: #888; margin: 4px 0 0; font-size: 13px;">Admin notification</p>
  </div>
  <div style="padding: 30px 0;">
    <h2 style="color: #231F1F; margin-top: 0;">%s</h2>
    <p style="color: #555; line-height: 1.6;">%s</p>
  </div>
</body></html>`, headline, body)

	return s.notification.Send(toEmail, subject, html)
}

// SendAuthorPostStatus is the generic "your post changed status" mail to an author.
func (s *EmailService) SendAuthorPostStatus(toEmail, authorName, headline, body, postTitle, postSlug string) error {
	if !s.notification.enabled {
		return nil
	}

	postURL := fmt.Sprintf("%s/blog/%s", siteURL(), postSlug)
	subject := fmt.Sprintf("%s — %s", postTitle, headline)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; padding: 20px 0; border-bottom: 2px solid #FF9C00;">
    <h1 style="color: #231F1F; margin: 0;">247 Tech Spyware</h1>
  </div>
  <div style="padding: 30px 0;">
    <p style="color: #231F1F; font-size: 16px;">Hi %s,</p>
    <h2 style="color: #231F1F;">%s</h2>
    <p style="color: #555; line-height: 1.6;">%s</p>
    <div style="text-align: center; padding: 20px 0;">
      <a href="%s" style="background-color: #FF9C00; color: #231F1F; padding: 12px 30px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">View Post</a>
    </div>
  </div>
</body></html>`, authorName, headline, body, postURL)

	return s.notification.Send(toEmail, subject, html)
}
