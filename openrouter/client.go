package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const BaseURL = "https://openrouter.ai/api/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Response struct {
	Choices []Choice `json:"choices"`
}

type Client struct {
	Model       string
	APIKey      string
	Timeout     int
	Temperature float64
}

func NewClient(model, apiKey string, timeout int, temperature float64) *Client {
	return &Client{
		Model:       model,
		APIKey:      apiKey,
		Timeout:     timeout,
		Temperature: temperature,
	}
}

func (c *Client) Chat(messages []Message) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OpenRouter API key no configurada. Exportá OPENROUTER_API_KEY=sk-or-...")
	}

	reqBody := Request{
		Model:       c.Model,
		Messages:    messages,
		Temperature: c.Temperature,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{Timeout: time.Duration(c.Timeout) * time.Second}
	req, err := http.NewRequest("POST", BaseURL, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("HTTP-Referer", "https://github.com/juano/deyaclaw")
	req.Header.Set("X-Title", "DeyaClaw")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error conectando a OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
    	var errBody bytes.Buffer
    	errBody.ReadFrom(resp.Body)
    		return "", fmt.Errorf("OpenRouter respondió con status %d: %s", resp.StatusCode, errBody.String())
	}


	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decodificando respuesta: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("OpenRouter no devolvió respuestas")
	}

	return result.Choices[0].Message.Content, nil
}
