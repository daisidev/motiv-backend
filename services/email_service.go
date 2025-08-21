package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"

	"github.com/hidenkeys/motiv-backend/models"
)

// Helper function for safe string truncation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type EmailService interface {
	SendTicketConfirmation(ticket *models.Ticket, event *models.Event, user *models.User) error
	SendHostNotification(ticket *models.Ticket, event *models.Event, user *models.User, host *models.User) error
	SendPasswordResetEmail(user *models.User, resetToken string) error
	SendWelcomeEmail(user *models.User) error
}

type ZohoEmailService struct {
	fromEmail string
	password   string
	smtpHost   string
	smtpPort   string
}

func NewZohoEmailService() EmailService {
	fromEmail := "teniola.sobande@zidihq.com"
	password := "EVYvaG9HJGrh"

	log.Printf("=== EMAIL SERVICE INITIALIZATION ===")
	log.Printf("Using Zoho SMTP service")
	log.Printf("From email: %s", fromEmail)
	log.Printf("SMTP Host: smtp.zoho.com:587")
	log.Printf("======================================")

	return &ZohoEmailService{
		fromEmail: fromEmail,
		password:  password,
		smtpHost:  "smtp.zoho.com",
		smtpPort:  "587",
	}
}

func (e *ZohoEmailService) SendTicketConfirmation(ticket *models.Ticket, event *models.Event, user *models.User) error {
	log.Printf("=== SENDING TICKET CONFIRMATION EMAIL ===")
	log.Printf("Ticket ID: %s", ticket.ID.String())
	log.Printf("Event: %s", event.Title)
	log.Printf("User: %s (%s)", user.Name, user.Email)
	log.Printf("Attendee: %s (%s)", ticket.AttendeeFullName, ticket.AttendeeEmail)

	subject := fmt.Sprintf("Your Ticket for %s - Confirmation", event.Title)
	log.Printf("Email subject: %s", subject)

	htmlContent, _, err := e.generateTicketConfirmationContent(ticket, event, user)
	if err != nil {
		log.Printf("❌ Failed to generate ticket confirmation content: %v", err)
		return fmt.Errorf("failed to generate email content: %w", err)
	}
	log.Printf("✅ Email content generated successfully")

	err = e.sendEmail(ticket.AttendeeEmail, subject, htmlContent)
	if err != nil {
		log.Printf("❌ Failed to send ticket confirmation email: %v", err)
		return err
	}
	log.Printf("✅ Ticket confirmation email sent successfully")
	log.Printf("==========================================")

	return nil
}

func (e *ZohoEmailService) SendHostNotification(ticket *models.Ticket, event *models.Event, user *models.User, host *models.User) error {
	log.Printf("=== SENDING HOST NOTIFICATION EMAIL ===")
	log.Printf("Host: %s (%s)", host.Name, host.Email)
	log.Printf("Event: %s (ID: %d)", event.Title, event.ID)
	log.Printf("User: %s (%s)", user.Name, user.Email)
	log.Printf("Ticket ID: %d", ticket.ID)

	subject := fmt.Sprintf("New Ticket Purchase for %s", event.Title)

	htmlContent, _, err := e.generateHostNotificationContent(ticket, event, user, host)
	if err != nil {
		log.Printf("❌ Failed to generate host notification content: %v", err)
		return fmt.Errorf("failed to generate email content: %w", err)
	}
	log.Printf("✅ Host notification content generated successfully")

	log.Printf("📧 Calling sendEmail for host notification...")
	err = e.sendEmail(host.Email, subject, htmlContent)
	if err != nil {
		log.Printf("❌ Host notification failed: %v", err)
		return err
	}
	log.Printf("✅ HOST NOTIFICATION EMAIL SENT SUCCESSFULLY!")
	log.Printf("====================================")
	return nil
}

func (e *ZohoEmailService) SendPasswordResetEmail(user *models.User, resetToken string) error {
	subject := "Reset Your Password - Motiv Events"

	htmlContent, _, err := e.generatePasswordResetContent(user, resetToken)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	return e.sendEmail(user.Email, subject, htmlContent)
}

func (e *ZohoEmailService) SendWelcomeEmail(user *models.User) error {
	log.Printf("=== SENDING WELCOME EMAIL ===")
	log.Printf("User: %s (%s)", user.Name, user.Email)
	log.Printf("User ID: %d", user.ID)

	subject := "Welcome to Motiv Events!"

	htmlContent, _, err := e.generateWelcomeEmailContent(user)
	if err != nil {
		log.Printf("❌ Failed to generate welcome email content: %v", err)
		return fmt.Errorf("failed to generate email content: %w", err)
	}
	log.Printf("✅ Welcome email content generated successfully")

	log.Printf("📧 Calling sendEmail for welcome email...")
	err = e.sendEmail(user.Email, subject, htmlContent)
	if err != nil {
		log.Printf("❌ Welcome email failed: %v", err)
		return err
	}
	log.Printf("✅ WELCOME EMAIL SENT SUCCESSFULLY!")
	log.Printf("==============================")
	return nil
}

func (e *ZohoEmailService) sendEmail(to, subject, body string) error {
	log.Printf("=== ZOHO SMTP EMAIL SENDING ===")
	log.Printf("To: %s", to)
	log.Printf("Subject: %s", subject)
	log.Printf("From: %s", e.fromEmail)
	log.Printf("SMTP Host: %s:%s", e.smtpHost, e.smtpPort)

	// Use Zoho SMTP host in PlainAuth
	auth := smtp.PlainAuth("", e.fromEmail, e.password, e.smtpHost)

	msg := "From: " + e.fromEmail + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n\n" +
		body

	err := smtp.SendMail(e.smtpHost+":"+e.smtpPort, auth, e.fromEmail, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("❌ Failed to send email via SMTP: %v", err)
		return fmt.Errorf("error sending email: %w", err)
	}
	
	log.Printf("✅ EMAIL SENT SUCCESSFULLY VIA ZOHO SMTP!")
	log.Printf("===============================")
	return nil
}

func (e *ZohoEmailService) generateTicketConfirmationContent(ticket *models.Ticket, event *models.Event, user *models.User) (string, string, error) {
	// HTML Template
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ticket Confirmation</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 10px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { background: #667eea; color: white; padding: 20px; text-align: center; border-radius: 10px 10px 0 0; margin: -20px -20px 20px -20px; }
        .ticket-info { background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .qr-code { text-align: center; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #666; }
        .btn { display: inline-block; background: #667eea; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Ticket Confirmed!</h1>
            <p>Your ticket for {{.Event.Title}} has been confirmed</p>
        </div>
        
        <h2>Hi {{.Ticket.AttendeeFullName}},</h2>
        <p>Thank you for your purchase! Here are your ticket details:</p>
        
        <div class="ticket-info">
            <h3>📅 Event Details</h3>
            <p><strong>Event:</strong> {{.Event.Title}}</p>
            <p><strong>Date:</strong> {{.Event.StartDate.Format "Monday, January 2, 2006"}}</p>
            <p><strong>Time:</strong> {{.Event.StartTime}} - {{.Event.EndTime}}</p>
            <p><strong>Location:</strong> {{.Event.Location}}</p>
            {{if .Event.Description}}
            <p><strong>Description:</strong> {{.Event.Description}}</p>
            {{end}}
            {{if .Event.ManualDescription}}
            <div style="background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 5px; padding: 10px; margin: 10px 0;">
                <p><strong>⚠️ Important Information:</strong></p>
                <p style="margin: 5px 0;">{{.Event.ManualDescription}}</p>
            </div>
            {{end}}
        </div>
        
        <div class="ticket-info">
            <h3>🎫 Ticket Information</h3>
            <p><strong>Ticket ID:</strong> {{.Ticket.ID}}</p>
            <p><strong>Attendee:</strong> {{.Ticket.AttendeeFullName}}</p>
            <p><strong>Email:</strong> {{.Ticket.AttendeeEmail}}</p>
            {{if .Ticket.AttendeePhone}}
            <p><strong>Phone:</strong> {{.Ticket.AttendeePhone}}</p>
            {{end}}
            {{if .Ticket.PaymentReference}}
            <p><strong>Payment Reference:</strong> {{.Ticket.PaymentReference}}</p>
            {{end}}
        </div>
        
        {{if .Ticket.QRCodeData}}
        <div class="qr-code">
            <h3>📱 Your QR Code</h3>
            <p>Show this QR code at the event entrance:</p>
            <img src="data:image/png;base64,{{.Ticket.QRCodeData}}" alt="QR Code" style="max-width: 200px;">
        </div>
        {{end}}
        
        <div style="margin: 30px 0; text-align: center;">
            <a href="{{.AppURL}}/my-raves" class="btn">View My Tickets</a>
        </div>
        
        <div class="footer">
            <p>Need help? Contact us at support@motivevents.com</p>
            <p>© 2025 Motiv Events. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	// Text Template (fallback)
	textTemplate := `
Ticket Confirmation - {{.Event.Title}}

Hi {{.Ticket.AttendeeFullName}},

Thank you for your purchase! Here are your ticket details:

EVENT DETAILS
Event: {{.Event.Title}}
Date: {{.Event.StartDate.Format "Monday, January 2, 2006"}}
Time: {{.Event.StartTime}} - {{.Event.EndTime}}
Location: {{.Event.Location}}
{{if .Event.Description}}Description: {{.Event.Description}}{{end}}
{{if .Event.ManualDescription}}

⚠️ IMPORTANT INFORMATION:
{{.Event.ManualDescription}}
{{end}}

TICKET INFORMATION
Ticket ID: {{.Ticket.ID}}
Attendee: {{.Ticket.AttendeeFullName}}
Email: {{.Ticket.AttendeeEmail}}
{{if .Ticket.AttendeePhone}}Phone: {{.Ticket.AttendeePhone}}{{end}}
{{if .Ticket.PaymentReference}}Payment Reference: {{.Ticket.PaymentReference}}{{end}}

Please save this email and bring your QR code to the event.

Need help? Contact us at support@motivevents.com

© 2025 Motiv Events. All rights reserved.
`

	data := struct {
		Ticket *models.Ticket
		Event  *models.Event
		User   *models.User
		AppURL string
	}{
		Ticket: ticket,
		Event:  event,
		User:   user,
		AppURL: os.Getenv("FRONTEND_URL"),
	}

	// Generate HTML content
	htmlTmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return "", "", err
	}
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	// Generate text content
	textTmpl, err := template.New("text").Parse(textTemplate)
	if err != nil {
		return "", "", err
	}
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", err
	}

	return htmlBuf.String(), textBuf.String(), nil
}

func (e *ZohoEmailService) generateHostNotificationContent(ticket *models.Ticket, event *models.Event, user *models.User, host *models.User) (string, string, error) {
	// HTML Template for host notification
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Ticket Purchase</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 10px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { background: #28a745; color: white; padding: 20px; text-align: center; border-radius: 10px 10px 0 0; margin: -20px -20px 20px -20px; }
        .ticket-info { background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #666; }
        .btn { display: inline-block; background: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; margin: 10px 0; }
        .highlight { background: #e8f5e8; padding: 10px; border-left: 4px solid #28a745; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>💰 New Ticket Sale!</h1>
            <p>Someone just purchased a ticket for your event</p>
        </div>
        
        <h2>Hi {{.Host.Name}},</h2>
        <p>Great news! You have a new ticket purchase for your event:</p>
        
        <div class="highlight">
            <h3>🎉 {{.Event.Title}}</h3>
        </div>
        
        <div class="ticket-info">
            <h3>👤 Customer Details</h3>
            <p><strong>Name:</strong> {{.Ticket.AttendeeFullName}}</p>
            <p><strong>Email:</strong> {{.Ticket.AttendeeEmail}}</p>
            {{if .Ticket.AttendeePhone}}
            <p><strong>Phone:</strong> {{.Ticket.AttendeePhone}}</p>
            {{end}}
            {{if .Ticket.PaymentReference}}
            <p><strong>Payment Reference:</strong> {{.Ticket.PaymentReference}}</p>
            {{end}}
        </div>
        
        <div class="ticket-info">
            <h3>📅 Event Details</h3>
            <p><strong>Date:</strong> {{.Event.StartDate.Format "Monday, January 2, 2006"}}</p>
            <p><strong>Time:</strong> {{.Event.StartTime}} - {{.Event.EndTime}}</p>
            <p><strong>Location:</strong> {{.Event.Location}}</p>
        </div>
        
        <div style="margin: 30px 0; text-align: center;">
            <a href="{{.AppURL}}/hosts/events/{{.Event.ID}}" class="btn">View Event Dashboard</a>
        </div>
        
        <div class="footer">
            <p>Keep up the great work! Your event is gaining traction.</p>
            <p>© 2025 Motiv Events. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	// Text Template for host notification
	textTemplate := `
New Ticket Purchase - {{.Event.Title}}

Hi {{.Host.Name}},

Great news! You have a new ticket purchase for your event: {{.Event.Title}}

CUSTOMER DETAILS
Name: {{.Ticket.AttendeeFullName}}
Email: {{.Ticket.AttendeeEmail}}
{{if .Ticket.AttendeePhone}}Phone: {{.Ticket.AttendeePhone}}{{end}}
{{if .Ticket.PaymentReference}}Payment Reference: {{.Ticket.PaymentReference}}{{end}}

EVENT DETAILS
Date: {{.Event.StartDate.Format "Monday, January 2, 2006"}}
Time: {{.Event.StartTime}} - {{.Event.EndTime}}
Location: {{.Event.Location}}

Keep up the great work! Your event is gaining traction.

© 2025 Motiv Events. All rights reserved.
`

	data := struct {
		Ticket *models.Ticket
		Event  *models.Event
		User   *models.User
		Host   *models.User
		AppURL string
	}{
		Ticket: ticket,
		Event:  event,
		User:   user,
		Host:   host,
		AppURL: os.Getenv("FRONTEND_URL"),
	}

	// Generate HTML content
	htmlTmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return "", "", err
	}
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	// Generate text content
	textTmpl, err := template.New("text").Parse(textTemplate)
	if err != nil {
		return "", "", err
	}
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", err
	}

	return htmlBuf.String(), textBuf.String(), nil
}

func (e *ZohoEmailService) generatePasswordResetContent(user *models.User, resetToken string) (string, string, error) {
	// HTML Template for password reset
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 10px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { background: #D72638; color: white; padding: 20px; text-align: center; border-radius: 10px 10px 0 0; margin: -20px -20px 20px -20px; }
        .content { padding: 20px 0; }
        .reset-info { background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #D72638; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #666; }
        .btn { display: inline-block; background: #D72638; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; font-weight: bold; }
        .btn:hover { background: #B91E2F; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; color: #856404; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Reset Your Password</h1>
            <p>We received a request to reset your password</p>
        </div>
        
        <div class="content">
            <h2>Hi {{.User.Name}},</h2>
            <p>You recently requested to reset your password for your Motiv Events account. Click the button below to reset it:</p>
            
            <div style="text-align: center; margin: 30px 0;">
                <a href="{{.AppURL}}/reset-password?token={{.ResetToken}}" class="btn">Reset My Password</a>
            </div>
            
            <div class="reset-info">
                <h3>🔒 Security Information</h3>
                <p><strong>This link will expire in 1 hour</strong> for your security.</p>
                <p>If you didn't request this password reset, you can safely ignore this email. Your password will remain unchanged.</p>
            </div>
            
            <div class="warning">
                <p><strong>⚠️ Important:</strong> If the button above doesn't work, copy and paste this link into your browser:</p>
                <p style="word-break: break-all; font-family: monospace; background: #f8f9fa; padding: 10px; border-radius: 3px;">
                    {{.AppURL}}/reset-password?token={{.ResetToken}}
                </p>
            </div>
        </div>
        
        <div class="footer">
            <p>If you have any questions, contact us at support@motivevents.com</p>
            <p>© 2025 Motiv Events. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	// Text Template for password reset
	textTemplate := `
Reset Your Password - Motiv Events

Hi {{.User.Name}},

You recently requested to reset your password for your Motiv Events account.

To reset your password, click the following link:
{{.AppURL}}/reset-password?token={{.ResetToken}}

SECURITY INFORMATION:
- This link will expire in 1 hour for your security
- If you didn't request this password reset, you can safely ignore this email
- Your password will remain unchanged if you don't click the link

If you have any questions, contact us at support@motivevents.com

© 2025 Motiv Events. All rights reserved.
`

	data := struct {
		User       *models.User
		ResetToken string
		AppURL     string
	}{
		User:       user,
		ResetToken: resetToken,
		AppURL:     os.Getenv("FRONTEND_URL"),
	}

	// Generate HTML content
	htmlTmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return "", "", err
	}
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	// Generate text content
	textTmpl, err := template.New("text").Parse(textTemplate)
	if err != nil {
		return "", "", err
	}
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", err
	}

	return htmlBuf.String(), textBuf.String(), nil
}

func (e *ZohoEmailService) generateWelcomeEmailContent(user *models.User) (string, string, error) {
	// HTML Template
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to Motiv Events</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 28px; font-weight: bold; color: #6366f1; margin-bottom: 10px; }
        .welcome-title { font-size: 24px; color: #333; margin-bottom: 20px; }
        .content { color: #666; line-height: 1.6; margin-bottom: 30px; }
        .highlight { background-color: #f8fafc; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #6366f1; }
        .features { margin: 20px 0; }
        .feature { margin: 10px 0; padding-left: 20px; position: relative; }
        .feature:before { content: "✓"; position: absolute; left: 0; color: #10b981; font-weight: bold; }
        .cta { text-align: center; margin: 30px 0; }
        .button { display: inline-block; background-color: #6366f1; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; font-weight: bold; }
        .footer { text-align: center; color: #999; font-size: 14px; margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">Motiv Events</div>
            <h1 class="welcome-title">Welcome to Motiv, {{.Name}}! 🎉</h1>
        </div>
        
        <div class="content">
            <p>We're excited to have you join our community of event enthusiasts! Whether you're looking to discover amazing events or host your own, Motiv is here to make it happen.</p>
            
            <div class="highlight">
                <strong>Your Account Details:</strong><br>
                <strong>Name:</strong> {{.Name}}<br>
                <strong>Username:</strong> {{.Username}}<br>
                <strong>Email:</strong> {{.Email}}
            </div>
            
            <p>Here's what you can do with Motiv:</p>
            
            <div class="features">
                <div class="feature">Discover exciting events in your area</div>
                <div class="feature">Purchase tickets securely</div>
                <div class="feature">Create and host your own events</div>
                <div class="feature">Connect with like-minded people</div>
                <div class="feature">Get personalized event recommendations</div>
            </div>
            
            <p>Ready to explore? Start by browsing events in your area or create your first event!</p>
            
            <div class="cta">
                <a href="{{.AppURL}}" class="button">Explore Events</a>
            </div>
        </div>
        
        <div class="footer">
            <p>Thanks for joining Motiv Events!<br>
            If you have any questions, feel free to reach out to our support team.</p>
            <p>© 2025 Motiv Events. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	// Text Template
	textTemplate := `Welcome to Motiv Events, {{.Name}}!

We're excited to have you join our community of event enthusiasts! Whether you're looking to discover amazing events or host your own, Motiv is here to make it happen.

Your Account Details:
Name: {{.Name}}
Username: {{.Username}}
Email: {{.Email}}

Here's what you can do with Motiv:
✓ Discover exciting events in your area
✓ Purchase tickets securely
✓ Create and host your own events
✓ Connect with like-minded people
✓ Get personalized event recommendations

Ready to explore? Start by browsing events in your area or create your first event!

Visit: {{.AppURL}}

Thanks for joining Motiv Events!
If you have any questions, feel free to reach out to our support team.

© 2025 Motiv Events. All rights reserved.`

	// Get app URL from environment or use default
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://motiv.events"
	}

	// Data for template
	data := struct {
		Name     string
		Username string
		Email    string
		AppURL   string
	}{
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		AppURL:   appURL,
	}

	// Generate HTML content
	htmlTmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return "", "", err
	}
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	// Generate text content
	textTmpl, err := template.New("text").Parse(textTemplate)
	if err != nil {
		return "", "", err
	}
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", err
	}

	return htmlBuf.String(), textBuf.String(), nil
}
