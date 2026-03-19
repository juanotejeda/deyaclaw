package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"os"
	"path/filepath"
)

type CVESearchTool struct{}

func (t *CVESearchTool) Name() string { return "cvesearch" }
func (t *CVESearchTool) Description() string {
	return "Busca CVEs en NVD. Params: query (uno o varios separados por coma, ej: 'openssh 9.2,samba 4.6'), severity (opcional: LOW,MEDIUM,HIGH,CRITICAL)"
}
func loadMsfModules() map[string]string {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil
    }
    data, err := os.ReadFile(filepath.Join(home, ".deyaclaw", "msf_modules.json"))
    if err != nil {
        return nil
    }
    var m map[string]string
    if err := json.Unmarshal(data, &m); err != nil {
        return nil
    }
    return m
}

func (t *CVESearchTool) Execute(params map[string]string) Result {
	query := params["query"]
	if query == "" {
		return Result{ToolName: t.Name(), Error: fmt.Errorf("parámetro 'query' requerido")}
	}
	msfModules := loadMsfModules()
	queries := strings.Split(query, ",")
	var sb strings.Builder
	totalFound := 0

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}

		cves, err := searchNVD(q, params["severity"], 3)
		if err != nil {
			sb.WriteString(fmt.Sprintf("⚠️  Error buscando '%s': %s\n\n", q, err))
			continue
		}
		if len(cves) == 0 {
			sb.WriteString(fmt.Sprintf("— Sin CVEs para: %s\n\n", q))
			continue
		}

		totalFound += len(cves)
		sb.WriteString(fmt.Sprintf("=== %s ===\n", strings.ToUpper(q)))
		for _, c := range cves {
			sb.WriteString(fmt.Sprintf("┌─ %s [%s | %.1f]\n", c.ID, c.Severity, c.Score))
			sb.WriteString(fmt.Sprintf("│  %s\n", c.Description))
			sb.WriteString(fmt.Sprintf("└─ https://nvd.nist.gov/vuln/detail/%s\n\n", c.ID))
			if msfModules != nil {
			    if msf, ok := msfModules[c.ID]; ok {
			        sb.WriteString(fmt.Sprintf("   🎯 MSF: %s\n", msf))
			            							}
								}
		}
	}

	if totalFound == 0 {
		return Result{ToolName: t.Name(), Output: "No se encontraron CVEs para los servicios indicados."}
	}

	return Result{ToolName: t.Name(), Output: sb.String()}
}

type cveEntry struct {
	ID          string
	Published   string
	Severity    string
	Score       float64
	Description string
}

func searchNVD(query, severity string, max int) ([]cveEntry, error) {
	base := "https://services.nvd.nist.gov/rest/json/cves/2.0"
	p := url.Values{}
	p.Set("keywordSearch", query)
	p.Set("resultsPerPage", "10")

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", base+"?"+p.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "DeyaClaw/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NVD status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	all, err := parseNVD(body)
	if err != nil {
		return nil, err
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Score > all[j].Score })

	if severity != "" {
		var filtered []cveEntry
		for _, c := range all {
			if strings.EqualFold(c.Severity, severity) {
				filtered = append(filtered, c)
				if len(filtered) >= max {
					break
				}
			}
		}
		return filtered, nil
	}

	if len(all) > max {
		return all[:max], nil
	}
	return all, nil
}

func parseNVD(data []byte) ([]cveEntry, error) {
	var raw struct {
		Vulnerabilities []struct {
			CVE struct {
				ID           string `json:"id"`
				Published    string `json:"published"`
				Descriptions []struct {
					Lang  string `json:"lang"`
					Value string `json:"value"`
				} `json:"descriptions"`
				Metrics struct {
					CVSSv3 []struct {
						CVSSData struct {
							BaseScore    float64 `json:"baseScore"`
							BaseSeverity string  `json:"baseSeverity"`
						} `json:"cvssData"`
					} `json:"cvssMetricV31"`
					CVSSv2 []struct {
						CVSSData struct {
							BaseScore float64 `json:"baseScore"`
						} `json:"cvssData"`
						BaseSeverity string `json:"baseSeverity"`
					} `json:"cvssMetricV2"`
				} `json:"metrics"`
			} `json:"cve"`
		} `json:"vulnerabilities"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var entries []cveEntry
	for _, v := range raw.Vulnerabilities {
		cve := v.CVE
		desc := ""
		for _, d := range cve.Descriptions {
			if d.Lang == "en" {
				desc = d.Value
				if len(desc) > 200 {
					desc = desc[:200] + "..."
				}
				break
			}
		}

		score, severity := 0.0, "N/A"
		if len(cve.Metrics.CVSSv3) > 0 {
			score = cve.Metrics.CVSSv3[0].CVSSData.BaseScore
			severity = cve.Metrics.CVSSv3[0].CVSSData.BaseSeverity
		} else if len(cve.Metrics.CVSSv2) > 0 {
			score = cve.Metrics.CVSSv2[0].CVSSData.BaseScore
			severity = cve.Metrics.CVSSv2[0].BaseSeverity
		}

		if severity == "N/A" || severity == "" {
			switch {
			case score >= 9.0:
				severity = "CRITICAL"
			case score >= 7.0:
				severity = "HIGH"
			case score >= 4.0:
				severity = "MEDIUM"
			case score > 0:
				severity = "LOW"
			}
		}

		published := cve.Published
		if len(published) > 10 {
			published = published[:10]
		}

		entries = append(entries, cveEntry{
			ID: cve.ID, Published: published,
			Severity: severity, Score: score, Description: desc,
		})
	}
	return entries, nil
}
