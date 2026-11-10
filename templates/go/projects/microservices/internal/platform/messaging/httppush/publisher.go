package httppush

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sample-app/internal/platform/messaging"
)

type Publisher struct {
	url    string
	client *http.Client
}

func NewPublisher(url string) *Publisher {
	return &Publisher{url: url, client: &http.Client{Timeout: 10 * time.Second}}
}

func (p *Publisher) Publish(ctx context.Context, e messaging.Envelope) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s respondeu %d", p.url, res.StatusCode)
	}
	return nil
}
