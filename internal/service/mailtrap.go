package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type MailTrapService struct {
	apiKey string
	url    string
	client *http.Client
	method string
}

type mailtrapAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type mailtrapPayload struct {
	From    mailtrapAddress   `json:"from"`
	To      []mailtrapAddress `json:"to"`
	Subject string            `json:"subject"`
	Text    string            `json:"text"`
}

func NewMailTrapService(apiKey string, url string) *MailTrapService {
	return &MailTrapService{
		apiKey: apiKey,
		url:    url,
		client: &http.Client{},
		method: http.MethodPost,
	}
}

func (s *MailTrapService) constructPayload(to string, subject string, body string) (*bytes.Buffer, error) {

	payload := mailtrapPayload{
		From: mailtrapAddress{
			Email: "mailtrap@example.com",
			Name:  "Mailtrap",
		},
		To: []mailtrapAddress{
			{Email: to},
		},
		Subject: subject,
		Text:    body,
	}

	js, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(js), nil
}

func (s *MailTrapService) SendEmail(to string, subject string, body string, messageId string) error {

	payload, err := s.constructPayload(to, subject, body)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(s.method, s.url, payload)

	if err != nil {
		slog.Error(
			"Failed to create HTTP request",
			"error", err,
		)
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Message-Id", messageId)

	res, err := s.client.Do(req)

	if err != nil {
		slog.Error(
			"Failed to execute HTTP request",
			"method", req.Method,
			"error", err,
		)
		return err
	}
	defer res.Body.Close()

	resbody, err := io.ReadAll(res.Body)

	if err != nil {
		slog.Error(
			"Failed to read HTTP response",
			"error", err,
		)
		return err
	}

	slog.Info(string(resbody))

	return nil
}
