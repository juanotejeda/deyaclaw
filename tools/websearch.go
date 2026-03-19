package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WebSearchTool struct{}

func (t *WebSearchTool) Name() string        { return "websearch" }
func (t *WebSearchTool) Description() string {
	return "Busca información en la web via DuckDuckGo. Útil para CVEs, exploits, técnicas, herramientas y documentación actualizada."
}
func (t *WebSearchTool) Params() []string { return []string{"query"} }

func (t *WebSearchTool) Execute(params map[string]string) Result {
	query := params["query"]
	if query == "" {
		return Result{
			ToolName: t.Name(),
			Error:    fmt.Errorf("parámetro 'query' requerido"),
		}
	}

	results, err := duckDuckGoSearch(query, 5)
	if err != nil {
		return Result{
			ToolName: t.Name(),
			Error:    fmt.Errorf("error en búsqueda: %w", err),
		}
	}

	if len(results) == 0 {
		return Result{
			ToolName: t.Name(),
			Output:   "No se encontraron resultados para: " + query,
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Resultados para: %s\n\n", query))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("    URL: %s\n", r.URL))
		sb.WriteString(fmt.Sprintf("    %s\n\n", r.Snippet))
	}

	return Result{
		ToolName: t.Name(),
		Output:   sb.String(),
	}
}

type searchResult struct {
	Title   string
	URL     string
	Snippet string
}

func duckDuckGoSearch(query string, maxResults int) ([]searchResult, error) {
	encodedQuery := url.QueryEscape(query)
	searchURL := "https://api.duckduckgo.com/?q=" + encodedQuery + "&format=json&no_html=1&skip_disambig=1"

	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "DeyaClaw/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseDDGJSON(body, maxResults)
}

func parseDDGJSON(data []byte, max int) ([]searchResult, error) {
	var raw struct {
		AbstractText   string `json:"AbstractText"`
		AbstractURL    string `json:"AbstractURL"`
		AbstractSource string `json:"AbstractSource"`
		RelatedTopics  []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
			Topics   []struct {
				Text     string `json:"Text"`
				FirstURL string `json:"FirstURL"`
			} `json:"Topics"`
		} `json:"RelatedTopics"`
		Results []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %w", err)
	}

	var results []searchResult

	// Abstract principal
	if raw.AbstractText != "" && raw.AbstractURL != "" {
		results = append(results, searchResult{
			Title:   raw.AbstractSource,
			URL:     raw.AbstractURL,
			Snippet: raw.AbstractText,
		})
	}

	// Results directos
	for _, r := range raw.Results {
		if len(results) >= max {
			break
		}
		if r.FirstURL == "" {
			continue
		}
		snippet := r.Text
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		results = append(results, searchResult{
			Title:   r.FirstURL,
			URL:     r.FirstURL,
			Snippet: snippet,
		})
	}

	// RelatedTopics como fallback
	for _, rt := range raw.RelatedTopics {
		if len(results) >= max {
			break
		}
		if rt.FirstURL != "" && rt.Text != "" {
			snippet := rt.Text
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}
			results = append(results, searchResult{
				Title:   rt.FirstURL,
				URL:     rt.FirstURL,
				Snippet: snippet,
			})
		}
		for _, sub := range rt.Topics {
			if len(results) >= max {
				break
			}
			if sub.FirstURL != "" {
				results = append(results, searchResult{
					Title:   sub.FirstURL,
					URL:     sub.FirstURL,
					Snippet: sub.Text,
				})
			}
		}
	}

	return results, nil
}
