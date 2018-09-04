package dropin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type SlackResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

func (sr *SlackResponse) bodyBuffer() (io.Reader, error) {
	b, err := json.Marshal(sr)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func (sr *SlackResponse) Send(URL string) error {
	body, err := sr.bodyBuffer()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", URL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
