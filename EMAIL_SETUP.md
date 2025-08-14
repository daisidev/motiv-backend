# Email Notifications Setup Guide

This guide explains how to set up email notifications for ticket purchases using Brevo (formerly SendinBlue) as the email service provider.

## Overview

The email service automatically sends two types of emails when a ticket is purchased:

1. **Ticket Confirmation Email** - Sent to the customer who purchased the ticket
2. **Host Notification Email** - Sent to the event host informing them of the new ticket sale

## Features

- ✅ Beautiful HTML email templates with embedded styling
- ✅ Fallback plain text versions for all emails
- ✅ QR code inclusion in ticket confirmation emails
- ✅ Event details and ticket information
- ✅ Automatic email sending for both paid tickets and free RSVPs
- ✅ Error handling that doesn't block ticket creation if emails fail

## Setup Instructions

### 1. Create a Brevo Account

1. Visit [brevo.com](https://www.brevo.com) and create an account
2. Verify your email address
3. Complete the account setup process

### 2. Get Your API Key

1. Log into your Brevo dashboard
2. Go to **Settings** → **API Keys**
3. Click **Generate a new API key**
4. Give it a name (e.g., "Motiv Backend")
5. Copy the generated API key

### 3. Verify Your Sender Email

1. In your Brevo dashboard, go to **Settings** → **Senders & IP**
2. Click **Add a new sender**
3. Enter your email address (e.g., `noreply@yourdomain.com`)
4. Follow the verification process

### 4. Configure Environment Variables

Add the following environment variables to your `.env` file:

```bash
# Brevo Email Configuration
BREVO_API_KEY=your_actual_brevo_api_key_here
BREVO_SENDER_EMAIL=noreply@yourdomain.com
FRONTEND_URL=http://localhost:3000
```

**Important Notes:**
- Replace `your_actual_brevo_api_key_here` with your actual API key from step 2
- Replace `noreply@yourdomain.com` with the email address you verified in step 3
- Update `FRONTEND_URL` to match your frontend application URL

### 5. Test the Setup

You can test the email functionality by:

1. Creating a test event
2. Purchasing a ticket (or doing a free RSVP)
3. Check your email inbox and the host's email inbox

## Email Templates

### Ticket Confirmation Email

Sent to customers when they purchase a ticket or RSVP for a free event.

**Includes:**
- Event details (title, date, time, location)
- Ticket information (ID, attendee details)
- QR code (if available)
- Link to view tickets in the app

### Host Notification Email

Sent to event hosts when someone purchases a ticket for their event.

**Includes:**
- Customer details (name, email, phone)
- Event information
- Payment reference
- Link to event dashboard

## API Integration

The email service is automatically called in the following scenarios:

### Paid Tickets (Paystack Webhook)
```go
// In payment_handler.go - handleSuccessfulPayment()
h.emailService.SendTicketConfirmation(ticket, eventDetails, user)
h.emailService.SendHostNotification(ticket, eventDetails, user, host)
```

### Free RSVPs
```go
// In ticket_handler.go - RSVPFreeEvent()
h.emailService.SendTicketConfirmation(ticket, eventDetails, userDetails)
h.emailService.SendHostNotification(ticket, eventDetails, userDetails, host)
```

### Payment Simulation (Testing)
```go
// In payment_handler.go - SimulatePaymentSuccess()
h.emailService.SendTicketConfirmation(ticket, eventDetails, userDetails)
h.emailService.SendHostNotification(ticket, eventDetails, userDetails, host)
```

## Error Handling

The email service includes robust error handling:

- Email failures are logged but don't prevent ticket creation
- Network issues or API failures won't block the user's ticket purchase
- All email operations are non-blocking

## Customization

### Modifying Email Templates

The email templates are defined in `services/email_service.go`. You can customize:

- HTML styling and layout
- Email content and messaging
- Brand colors and logos
- Template structure

### Adding New Email Types

To add new email types:

1. Add new methods to the `EmailService` interface
2. Implement the methods in `BrevoEmailService`
3. Create new template functions
4. Call the methods from appropriate handlers

## Production Considerations

### Security
- Keep your Brevo API key secure and never commit it to version control
- Use environment variables for all configuration
- Consider using secret management services in production

### Monitoring
- Monitor email delivery rates in your Brevo dashboard
- Set up alerts for failed email sends
- Log email operations for debugging

### Scalability
- Brevo's free plan includes 300 emails per day
- Consider upgrading to a paid plan for higher volumes
- Implement email queuing for high-volume scenarios

### Domain Setup
- For production, use a custom domain for sender emails
- Set up SPF, DKIM, and DMARC records for better deliverability
- Consider using a subdomain like `mail.yourdomain.com`

## Troubleshooting

### Common Issues

1. **"Invalid API Key" Error**
   - Verify your API key is correct in the `.env` file
   - Check that the API key has the necessary permissions

2. **"Sender Email Not Verified" Error**
   - Ensure you've verified the sender email in Brevo dashboard
   - Check that `BREVO_SENDER_EMAIL` matches the verified email

3. **Emails Not Being Sent**
   - Check the application logs for error messages
   - Verify your Brevo account is active and has email credits
   - Test the API key with a simple curl request

4. **Templates Not Rendering Correctly**
   - Check for Go template syntax errors
   - Ensure all required data is being passed to templates
   - Test with minimal template content first

### Testing API Key

You can test your Brevo API key with this curl command:

```bash
curl -X GET \
  'https://api.brevo.com/v3/account' \
  -H 'Accept: application/json' \
  -H 'api-key: YOUR_API_KEY_HERE'
```

## Support

- [Brevo Documentation](https://developers.brevo.com/)
- [Brevo API Reference](https://developers.brevo.com/reference)
- [Go Template Documentation](https://golang.org/pkg/text/template/)
