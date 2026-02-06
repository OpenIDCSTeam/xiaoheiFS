package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"xiaoheiplay/internal/usecase"
)

const fcmEndpoint = "https://fcm.googleapis.com/fcm/send"

type FCMSender struct {
	http *http.Client
}

func NewFCMSender() *FCMSender {
	return &FCMSender{
		http: &http.Client{Timeout: 8 * time.Second},
	}
}

func (s *FCMSender) Send(ctx context.Context, serverKey string, tokens []string, payload usecase.PushPayload) error {
	if serverKey == "" || len(tokens) == 0 {
		return nil
	}
	const batchSize = 500
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		if err := s.sendBatch(ctx, serverKey, tokens[i:end], payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *FCMSender) sendBatch(ctx context.Context, serverKey string, tokens []string, payload usecase.PushPayload) error {
	body := map[string]any{
		"registration_ids": tokens,
		"priority":         "high",
		"notification": map[string]any{
			"title": payload.Title,
			"body":  payload.Body,
		},
	}
	if len(payload.Data) > 0 {
		body["data"] = payload.Data
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fcmEndpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+serverKey)
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fcm send failed: status %d", resp.StatusCode)
	}
	return nil
}
