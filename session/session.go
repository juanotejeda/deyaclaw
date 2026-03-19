package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/juano/deyaclaw/ollama"
)

type Session struct {
	Name      string           `json:"name"`
	Profile   string           `json:"profile"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
	History   []ollama.Message `json:"history"`
}

func sessionPath(sessionsDir, name string) string {
	return filepath.Join(sessionsDir, name+".json")
}

func Save(sessionsDir, name, profile string, history []ollama.Message) error {
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return err
	}

	path := sessionPath(sessionsDir, name)
	now  := time.Now().Format("2006-01-02 15:04:05")

	s := &Session{
		Name:      name,
		Profile:   profile,
		UpdatedAt: now,
		History:   history,
	}

	// Si ya existe, preservamos el created_at
	if existing, err := Load(sessionsDir, name); err == nil {
		s.CreatedAt = existing.CreatedAt
	} else {
		s.CreatedAt = now
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Load(sessionsDir, name string) (*Session, error) {
	path := sessionPath(sessionsDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sesión '%s' no encontrada", name)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("error leyendo sesión: %w", err)
	}

	return &s, nil
}

func List(sessionsDir string) ([]Session, error) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []Session
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			name := strings.TrimSuffix(e.Name(), ".json")
			s, err := Load(sessionsDir, name)
			if err != nil {
				continue
			}
			sessions = append(sessions, *s)
		}
	}

	return sessions, nil
}

func Delete(sessionsDir, name string) error {
	path := sessionPath(sessionsDir, name)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("no se pudo borrar la sesión '%s'", name)
	}
	return nil
}
