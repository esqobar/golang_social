package mailer

import (
	"fmt"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	//apiKey    string
	client *sendgrid.Client
}

func NewSendgrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		//apiKey:    apiKey,
		client: client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	from := mail.NewEmail(FromName, m.fromEmail) // Use a readable sender name
	to := mail.NewEmail(username, email)

	// Parse template
	subject, body, err := parseTemplate(templateFile, data)
	log.Println("SUBJECT:", subject)
	log.Println("BODY:", body)
	if err != nil {
		return -1, fmt.Errorf("template parse error: %w", err)
	}

	message := mail.NewSingleEmail(from, subject, to, "", body)

	//// Sandbox mode controls real delivery
	//message.SetMailSettings(&mail.MailSettings{
	//	SandboxMode: &mail.Setting{
	//		Enable: &isSandbox,
	//	},
	//})

	var retryErr error

	for i := 0; i < maxRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			retryErr = err
			log.Printf("[SendGrid] Error sending email (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		return response.StatusCode, nil
	}

	return -1, fmt.Errorf("failed to send email after %d attempts: %v", maxRetries, retryErr)
}

//func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
//	from := mail.NewEmail(FromName, m.fromEmail)
//	to := mail.NewEmail(username, email)
//
//	// Parse template
//	subject, body, err := parseTemplate(templateFile, data)
//	if err != nil {
//		return -1, err
//	}
//
//	message := mail.NewSingleEmail(from, subject, to, "", body)
//
//	message.SetMailSettings(&mail.MailSettings{
//		SandboxMode: &mail.Setting{
//			Enable: &isSandbox,
//		},
//	})
//
//	var lastErr error
//
//	for i := 0; i < maxRetries; i++ {
//		response, err := m.client.Send(message)
//
//		if err != nil {
//			lastErr = err
//			log.Printf("sendgrid error (attempt %d/%d): %v", i+1, maxRetries, err)
//
//			// exponential backoff (1s, 2s, 4s)
//			time.Sleep(time.Second * time.Duration(1<<i))
//			continue
//		}
//
//		if response.StatusCode >= 400 {
//			lastErr = fmt.Errorf("status=%d body=%s", response.StatusCode, response.Body)
//
//			log.Printf("sendgrid failed (attempt %d/%d): %v",
//				i+1, maxRetries, lastErr)
//
//			time.Sleep(time.Second * time.Duration(1<<i))
//			continue
//		}
//
//		log.Printf("email sent successfully: status=%d", response.StatusCode)
//		return response.StatusCode, nil
//	}
//
//	return -1, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
//}
