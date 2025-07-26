package pkg

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"server/config"
	"strings"
	"time"
)

const OtpTemplate = `
    <html lang="en">
    <body>
      <div style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px; border-radius: 8px; text-align: center;">
        <h1 style="color: #333;">Verify Your Email</h1>
        <p style="font-size: 16px; color: #555;">Use the OTP below to verify your email:</p>
        <div style="font-size: 24px; font-weight: bold; color: #28a745; margin: 20px 0; padding: 10px; background-color: #fff; display: inline-block; border-radius: 5px; letter-spacing: 3px;">
          {{otp}}
        </div>
        <p style="font-size: 14px; color: #777; margin-top: 20px;">This OTP will expire in 3 minutes.</p>
        <div style="margin-top: 30px; font-size: 12px; color: #aaa; text-align: center;">
          <p>If you did not request this email, please ignore it.</p>
          <p>&copy; {{date}} Tryntel. All rights reserved.</p>
        </div>
      </div>
    </body>
  </html>`


type emailRequestPayload struct {
	SenderAddress string            `json:"senderAddress"`
	Content       emailContent      `json:"content"`
	Recipients    emailRecipients   `json:"recipients"`
	ReplyTo       []emailAddress    `json:"replyTo,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

type emailContent struct {
	Subject   string `json:"subject"`
	PlainText string `json:"plainText,omitempty"`
	HTML      string `json:"html"`
}

type emailRecipients struct {
	To  []emailAddress `json:"to"`
	CC  []emailAddress `json:"cc,omitempty"`
	BCC []emailAddress `json:"bcc,omitempty"`
}

type emailAddress struct {
	Address     string `json:"address"`
	DisplayName string `json:"displayName,omitempty"`
}

// SendMail sends an email using Azure REST API.
func SendMail(recipientMail, templateName, subject, message string) error {

	cfg, err := config.SetupEnv()
	if err != nil {
		log.Fatalf("Failed to load env variables: %v", err)
	}
	if cfg.AzureCommunicationDNS == "" {
		return errors.New("AZURE_COMMUNICATION_DNS environment variable not set")
	}
	connectionString := cfg.AzureCommunicationDNS
	senderAddress := cfg.AzureEmailSenderAddress
	if senderAddress == "" {
		return errors.New("AZURE_SENDER_ADDRESS environment variable not set")
	}

	parts := strings.Split(connectionString, ";")
	var endpoint, accessKey string
	for _, part := range parts {
		if strings.HasPrefix(part, "endpoint=") {
			endpoint = strings.TrimPrefix(part, "endpoint=")
		} else if strings.HasPrefix(part, "accesskey=") {
			accessKey = strings.TrimPrefix(part, "accesskey=")
		}
	}
	if endpoint == "" || accessKey == "" {
		return errors.New("invalid connection string format")
	}

	var htmlContent string
	switch templateName {
	case "otp":
		tempContent := strings.Replace(OtpTemplate, "{{otp}}", message, 1)
		tempContent = strings.Replace(tempContent, "{{date}}", time.Now().Format("2006-01-02"), 1)
		htmlContent = tempContent
	case "alert":
		htmlContent = message
	default:
		return fmt.Errorf("unknown template name: %s", templateName)
	}

	payload := emailRequestPayload{
		SenderAddress: senderAddress,
		Content: emailContent{
			Subject: subject,
			HTML:    htmlContent,
		},
		Recipients: emailRecipients{
			To: []emailAddress{{Address: recipientMail}},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	apiPath := "/emails:send?api-version=2023-03-31"
	fullURL := endpoint + apiPath

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC1123)
	timestamp = strings.Replace(timestamp, "UTC", "GMT", 1)

	// Hash the payload
	hasher := sha256.New()
	hasher.Write(payloadBytes)
	contentHash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	// Get host from endpoint
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse endpoint URL: %w", err)
	}
	host := parsedURL.Host
	stringToSign := fmt.Sprintf("POST\n%s\n%s;%s;%s", apiPath, timestamp, host, contentHash)

	// Decode the access key and create the signature
	decodedAccessKey, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil {
		return fmt.Errorf("failed to decode access key: %w", err)
	}
	mac := hmac.New(sha256.New, decodedAccessKey)
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ms-date", timestamp)
	req.Header.Set("x-ms-content-sha256", contentHash)
	req.Header.Set("Authorization", "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature="+signature)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("Successfully sent '%s' email to %s via REST API", templateName, recipientMail)
	return nil
}
