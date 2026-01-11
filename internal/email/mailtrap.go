package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// MailtrapClient is a client for sending emails via Mailtrap
type MailtrapClient struct {
	apiToken   string
	senderEmail string
	senderName string
	httpClient *http.Client
}

// NewMailtrapClient creates a new Mailtrap client
func NewMailtrapClient(apiToken, senderEmail, senderName string) *MailtrapClient {
	return &MailtrapClient{
		apiToken:   apiToken,
		senderEmail: senderEmail,
		senderName: senderName,
		httpClient: &http.Client{},
	}
}

// EmailAddress represents an email address
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// SendEmailRequest is the request body for sending an email
type SendEmailRequest struct {
	From     EmailAddress   `json:"from"`
	To       []EmailAddress `json:"to"`
	Subject  string         `json:"subject"`
	Text     string         `json:"text,omitempty"`
	HTML     string         `json:"html,omitempty"`
	Category string         `json:"category,omitempty"`
}

// SendOTPEmail sends an OTP verification email
func (c *MailtrapClient) SendOTPEmail(to, otpCode string) error {
	subject := "Your Zyndra Verification Code"
	
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f8f8f8;">
  <div style="max-width: 480px; margin: 40px auto; background: white; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);">
    <!-- Header -->
    <div style="background: linear-gradient(135deg, #4F46E5 0%%, #06B6D4 100%%); padding: 32px; text-align: center;">
      <div style="width: 56px; height: 56px; background: white; border-radius: 12px; margin: 0 auto 16px; display: flex; align-items: center; justify-content: center;">
        <div style="color: #4F46E5; font-size: 24px; font-weight: bold;">Z</div>
      </div>
      <h1 style="color: white; margin: 0; font-size: 20px; font-weight: 600;">Zyndra</h1>
    </div>
    
    <!-- Content -->
    <div style="padding: 32px;">
      <h2 style="color: #1f2937; margin: 0 0 8px; font-size: 24px; font-weight: 600; text-align: center;">
        Verify your email
      </h2>
      <p style="color: #6b7280; margin: 0 0 24px; font-size: 14px; text-align: center; line-height: 1.5;">
        Enter this code to complete your registration
      </p>
      
      <!-- OTP Code -->
      <div style="background: #f3f4f6; border-radius: 12px; padding: 24px; text-align: center; margin-bottom: 24px;">
        <div style="font-size: 36px; font-weight: 700; letter-spacing: 8px; color: #1f2937; font-family: monospace;">
          %s
        </div>
      </div>
      
      <p style="color: #9ca3af; font-size: 12px; text-align: center; margin: 0;">
        This code expires in 10 minutes.<br>
        If you didn't request this code, you can safely ignore this email.
      </p>
    </div>
    
    <!-- Footer -->
    <div style="background: #f9fafb; padding: 16px 32px; text-align: center; border-top: 1px solid #e5e7eb;">
      <p style="color: #9ca3af; font-size: 12px; margin: 0;">
        © 2024 Zyndra. All rights reserved.
      </p>
    </div>
  </div>
</body>
</html>
`, otpCode)

	text := fmt.Sprintf(`Your Zyndra Verification Code

Your verification code is: %s

This code expires in 10 minutes.

If you didn't request this code, you can safely ignore this email.

© 2024 Zyndra. All rights reserved.
`, otpCode)

	return c.SendEmail(to, subject, text, html, "otp-verification")
}

// SendEmail sends an email via Mailtrap
func (c *MailtrapClient) SendEmail(to, subject, text, html, category string) error {
	if c.apiToken == "" {
		return fmt.Errorf("mailtrap API token not configured")
	}

	req := SendEmailRequest{
		From: EmailAddress{
			Email: c.senderEmail,
			Name:  c.senderName,
		},
		To: []EmailAddress{
			{Email: to},
		},
		Subject:  subject,
		Text:     text,
		HTML:     html,
		Category: category,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://send.api.mailtrap.io/api/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("mailtrap error (%d): %v", resp.StatusCode, errorResp)
		}
		return fmt.Errorf("mailtrap error: status %d", resp.StatusCode)
	}

	return nil
}

