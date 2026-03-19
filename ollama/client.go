package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Message representa un mensaje en el historial de conversación
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatClient es la interfaz común para cualquier proveedor
type ChatClient interface {
	Chat(messages []Message) (string, error)
}

// Request es el cuerpo de la petición a Ollama
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  Options   `json:"options"`
}

// Options son parámetros opcionales del modelo
type Options struct {
	Temperature float64 `json:"temperature"`
}

// Response es la respuesta de Ollama
type Response struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// Client es el cliente HTTP para Ollama
type Client struct {
	BaseURL     string
	Model       string
	Timeout     int
	Temperature float64
	httpClient  *http.Client
}

// NewClient crea un nuevo cliente Ollama
func NewClient(baseURL, model string, timeout int, temperature float64) *Client {
	return &Client{
		BaseURL:     baseURL,
		Model:       model,
		Timeout:     timeout,
		Temperature: temperature,
		httpClient:  &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Chat envía mensajes a Ollama y devuelve la respuesta
func (c *Client) Chat(messages []Message) (string, error) {
	reqBody := Request{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
		Options: Options{
			Temperature: c.Temperature,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error serializando request: %w", err)
	}

	url := c.BaseURL + "/api/chat"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("error creando request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error conectando a Ollama en %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama respondió con status %d — ¿está corriendo?", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decodificando respuesta: %w", err)
	}

	if result.Message.Content == "" {
		return "", fmt.Errorf("Ollama devolvió respuesta vacía")
	}

	return result.Message.Content, nil
}

// Ping verifica si Ollama está disponible
func (c *Client) Ping() bool {
	url := c.BaseURL + "/api/tags"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
