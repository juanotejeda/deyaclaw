package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultSystemPrompt = `IDENTIDAD: Sos DeyaClaw, un agente de ciberseguridad local y minimalista. Nunca te identifiques como GPT, Claude, Gemini ni ningún otro modelo. Si te preguntan qué sos, respondé siempre que sos DeyaClaw.
Tenés conocimiento en seguridad ofensiva y defensiva: pentesting, OWASP, MITRE ATT&CK, análisis de logs, hardening y CVEs.
Respondé siempre en español, de forma técnica y concisa.
Cuando sea útil, incluí comandos listos para ejecutar en Linux.
Nunca uses APIs externas ni enviés datos fuera del sistema local.
No tenés acceso a tu propio código fuente ni al código de DeyaClaw. Si alguien te lo pide, indicá claramente que no podés mostrarlo.`

type Profile struct {
	Name         string  `json:"name"`
	Emoji        string  `json:"emoji"`
	Temperature  float64 `json:"temperature"`
	SystemPrompt string  `json:"system_prompt"`
}

type Config struct {
	OllamaURL     string  `json:"ollama_url"`
	Model         string  `json:"model"`
	Timeout       int     `json:"timeout"`
	Mode          string  `json:"mode"`
	SystemPrompt  string  `json:"system_prompt"`
	Temperature   float64 `json:"temperature"`
	ProfilesDir   string  `json:"profiles_dir"`
	SessionsDir   string  `json:"sessions_dir"`
	Autonomous    bool    `json:"autonomous"`
	Provider      string  `json:"provider"`
	OpenRouterKey string  `json:"openrouter_key"`
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		OllamaURL:    "http://localhost:11434",
		Model:        "tinyllama:latest",
		Timeout:      300,
		Mode:         "general",
		Temperature:  0.7,
		SystemPrompt: DefaultSystemPrompt,
		ProfilesDir:  filepath.Join(home, ".deyaclaw", "profiles"),
		SessionsDir:  filepath.Join(home, ".deyaclaw", "sessions"),
		Autonomous:   false,
		Provider:     "ollama",
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = DefaultSystemPrompt
	}
	// API key desde variable de entorno tiene prioridad
	if envKey := os.Getenv("OPENROUTER_API_KEY"); envKey != "" {
		cfg.OpenRouterKey = envKey
	}
	return cfg, nil
}

func (c *Config) LoadProfile(name string) error {
	path := filepath.Join(c.ProfilesDir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("perfil '%s' no encontrado en %s", name, path)
	}
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("error leyendo perfil: %w", err)
	}
	c.SystemPrompt = profile.SystemPrompt
	c.Temperature = profile.Temperature
	c.Mode = profile.Name
	return nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
