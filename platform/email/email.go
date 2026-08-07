package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
	UseTLS    bool
}

type EmailService struct {
	config SMTPConfig
}

func NewEmailService(config SMTPConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	from := s.config.FromEmail
	fromDisplay := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		fromDisplay, to, subject, body)

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	if s.config.Port == 465 && s.config.UseTLS {
		return s.sendWithTLS(addr, auth, from, []string{to}, []byte(msg))
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func (s *EmailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host := strings.Split(addr, ":")[0]

	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail failed: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("smtp rcpt failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data failed: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	return client.Quit()
}

func (s *EmailService) SendModelCompletionEmail(to, userName, modelTitle string, status string) error {
	subject := fmt.Sprintf("Model Calculation Complete: %s", modelTitle)

	var statusColor, statusBgColor, statusText, statusIcon string
	if status == "completed" {
		statusColor = "#059669"
		statusBgColor = "#ecfdf5"
		statusText = "Successfully Completed"
		statusIcon = "✓"
	} else {
		statusColor = "#dc2626"
		statusBgColor = "#fef2f2"
		statusText = "Failed"
		statusIcon = "✕"
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;background-color:#f9fafb;-webkit-font-smoothing:antialiased">
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f9fafb;padding:40px 20px">
<tr><td align="center">
<table width="100%%" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#ffffff;border-radius:16px;box-shadow:0 4px 6px -1px rgba(0,0,0,0.1),0 2px 4px -2px rgba(0,0,0,0.1);overflow:hidden">

<!-- Header -->
<tr><td style="background-color:#111827;padding:32px 32px 28px 32px;text-align:center">
<table cellpadding="0" cellspacing="0" style="margin:0 auto">
<tr>
<td style="width:40px;height:40px;background-color:#374151;border-radius:10px;text-align:center;vertical-align:middle">
<span style="color:#ffffff;font-size:20px;font-weight:700">S</span>
</td>
<td style="padding-left:12px">
<span style="color:#ffffff;font-size:20px;font-weight:600;letter-spacing:-0.025em">Storcito</span>
</td>
</tr>
</table>
<p style="color:#9ca3af;margin:12px 0 0 0;font-size:13px;font-weight:400">Wildfire Risk Assessment Platform</p>
</td></tr>

<!-- Status Badge -->
<tr><td style="padding:32px 32px 0 32px;text-align:center">
<table cellpadding="0" cellspacing="0" style="margin:0 auto">
<tr>
<td style="background-color:%s;border-radius:50px;padding:10px 20px">
<span style="color:%s;font-size:14px;font-weight:600">%s %s</span>
</td>
</tr>
</table>
</td></tr>

<!-- Content -->
<tr><td style="padding:28px 32px 32px 32px">
<p style="color:#111827;font-size:18px;font-weight:600;margin:0 0 8px 0;text-align:center">Hello %s,</p>
<p style="color:#6b7280;font-size:14px;line-height:1.6;margin:0 0 24px 0;text-align:center">
Your model calculation has finished processing.
</p>

<!-- Model Card -->
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f9fafb;border-radius:12px;border:1px solid #e5e7eb">
<tr><td style="padding:20px">
<p style="color:#6b7280;margin:0 0 6px 0;font-size:11px;text-transform:uppercase;letter-spacing:0.05em;font-weight:500">Model Name</p>
<p style="color:#111827;margin:0;font-size:16px;font-weight:600">%s</p>
</td></tr>
</table>

<p style="color:#6b7280;font-size:14px;line-height:1.6;margin:24px 0 0 0;text-align:center">
View your results in the dashboard.
</p>
</td></tr>

<!-- Footer -->
<tr><td style="background-color:#f9fafb;padding:20px 32px;border-top:1px solid #e5e7eb">
<p style="color:#9ca3af;font-size:12px;margin:0;text-align:center;line-height:1.5">
This is an automated notification from Storcito.<br>
Please do not reply to this email.
</p>
</td></tr>

</table>
</td></tr>
</table>
</body>
</html>`, statusBgColor, statusColor, statusIcon, statusText, userName, modelTitle)

	return s.SendEmail(to, subject, body)
}

// SendModelSharedEmail notifies a recipient that a model has been shared with them.
func (s *EmailService) SendModelSharedEmail(to, recipientName, modelTitle, sharedByName, permission string) error {
	subject := fmt.Sprintf("%s shared a model with you: %s", sharedByName, modelTitle)

	if permission == "" {
		permission = "view"
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;background-color:#f9fafb;-webkit-font-smoothing:antialiased">
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f9fafb;padding:40px 20px">
<tr><td align="center">
<table width="100%%" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#ffffff;border-radius:16px;box-shadow:0 4px 6px -1px rgba(0,0,0,0.1),0 2px 4px -2px rgba(0,0,0,0.1);overflow:hidden">

<!-- Header -->
<tr><td style="background-color:#111827;padding:32px 32px 28px 32px;text-align:center">
<table cellpadding="0" cellspacing="0" style="margin:0 auto">
<tr>
<td style="width:40px;height:40px;background-color:#374151;border-radius:10px;text-align:center;vertical-align:middle">
<span style="color:#ffffff;font-size:20px;font-weight:700">S</span>
</td>
<td style="padding-left:12px">
<span style="color:#ffffff;font-size:20px;font-weight:600;letter-spacing:-0.025em">Storcito</span>
</td>
</tr>
</table>
<p style="color:#9ca3af;margin:12px 0 0 0;font-size:13px;font-weight:400">Wildfire Risk Assessment Platform</p>
</td></tr>

<!-- Badge -->
<tr><td style="padding:32px 32px 0 32px;text-align:center">
<table cellpadding="0" cellspacing="0" style="margin:0 auto">
<tr>
<td style="background-color:#eef2ff;border-radius:50px;padding:10px 20px">
<span style="color:#4f46e5;font-size:14px;font-weight:600">🔗 Model Shared</span>
</td>
</tr>
</table>
</td></tr>

<!-- Content -->
<tr><td style="padding:28px 32px 32px 32px">
<p style="color:#111827;font-size:18px;font-weight:600;margin:0 0 8px 0;text-align:center">Hello %s,</p>
<p style="color:#6b7280;font-size:14px;line-height:1.6;margin:0 0 24px 0;text-align:center">
<strong>%s</strong> has shared a model with you.
</p>

<!-- Model Card -->
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f9fafb;border-radius:12px;border:1px solid #e5e7eb">
<tr><td style="padding:20px">
<p style="color:#6b7280;margin:0 0 6px 0;font-size:11px;text-transform:uppercase;letter-spacing:0.05em;font-weight:500">Model Name</p>
<p style="color:#111827;margin:0 0 14px 0;font-size:16px;font-weight:600">%s</p>
<p style="color:#6b7280;margin:0 0 6px 0;font-size:11px;text-transform:uppercase;letter-spacing:0.05em;font-weight:500">Access Level</p>
<p style="color:#111827;margin:0;font-size:14px;font-weight:600;text-transform:capitalize">%s</p>
</td></tr>
</table>

<p style="color:#6b7280;font-size:14px;line-height:1.6;margin:24px 0 0 0;text-align:center">
Sign in to the dashboard to view this model.
</p>
</td></tr>

<!-- Footer -->
<tr><td style="background-color:#f9fafb;padding:20px 32px;border-top:1px solid #e5e7eb">
<p style="color:#9ca3af;font-size:12px;margin:0;text-align:center;line-height:1.5">
This is an automated notification from Storcito.<br>
Please do not reply to this email.
</p>
</td></tr>

</table>
</td></tr>
</table>
</body>
</html>`, recipientName, sharedByName, modelTitle, permission)

	return s.SendEmail(to, subject, body)
}
