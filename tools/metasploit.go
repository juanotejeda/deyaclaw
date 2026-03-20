package tools

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// =====================
// MetasploitTool
// =====================

type MetasploitTool struct{}

func (m *MetasploitTool) Name() string { return "metasploit" }

func parseWritableShare(output string) string {
	for _, line := range strings.Split(output, "\n") {
		l := strings.ToLower(line)
		if strings.Contains(l, "read/write") || strings.Contains(l, "rw") {
			// Formato típico: [+] 172.18.0.2 - myshare READ/WRITE
			parts := strings.Fields(line)
			for i, p := range parts {
				if strings.EqualFold(p, "READ/WRITE") || strings.EqualFold(p, "RW") {
					if i > 0 {
						return strings.Trim(parts[i-1], "-|[]")
					}
				}
			}
		}
	}
	return ""
}


func (m *MetasploitTool) Description() string {
	return "Ejecuta módulos de Metasploit Framework usando msfconsole. " +
		"USA ESTA TOOL (no shell) para cualquier módulo de Metasploit. " +
		"Params: module (ej: auxiliary/scanner/ssh/ssh_version), " +
		"options (string: \"CLAVE=VALOR CLAVE2=VALOR2\", ej: \"RHOSTS=127.0.0.1 RPORT=22\"), " +
		"timeout (opcional, segundos, default 120)."
}

func (m *MetasploitTool) Execute(params map[string]string) Result {
	module := params["module"]
	if module == "" {
		return Result{ToolName: m.Name(), Error: fmt.Errorf("parámetro 'module' requerido")}
	}

	timeout := 120 * time.Second
	if t := params["timeout"]; t != "" {
		var secs int
		fmt.Sscanf(t, "%d", &secs)
		if secs > 0 {
			timeout = time.Duration(secs) * time.Second
		}
	}

	var script strings.Builder
	script.WriteString(fmt.Sprintf("use %s; ", module))

	rawOptions := params["options"]
	reversePayloads := []string{"reverse_tcp", "reverse_http", "reverse_https", "meterpreter"}
	needsLHOST := false
	for _, p := range reversePayloads {
	    if strings.Contains(rawOptions, p) {
	        needsLHOST = true
	        break
	    }
	}
	if strings.HasPrefix(module, "exploit/") && needsLHOST && !strings.Contains(rawOptions, "LHOST=") {
	    if lhost := detectLocalIP(); lhost != "" {
	        script.WriteString(fmt.Sprintf("set LHOST %s; ", lhost))
	    }
	}
	

	if rawOptions != "" {
		for _, opt := range strings.Fields(rawOptions) {
			parts := strings.SplitN(opt, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if key != "" && val != "" {
					script.WriteString(fmt.Sprintf("set %s %s; ", key, val))
				}
			}
		}
	}

	script.WriteString("run; exit")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "msfconsole", "-q", "-x", script.String())
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	fmt.Printf("=== MSF SCRIPT ===\n%s\n==================\n", script.String())

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Result{ToolName: m.Name(), Error: fmt.Errorf("timeout: módulo superó %s", timeout)}
		}
		if strings.Contains(out.String(), "Failed to load module") ||
			strings.Contains(out.String(), "Invalid module") {
			return Result{ToolName: m.Name(), Error: fmt.Errorf("módulo no encontrado: %s", module)}
		}
		return Result{ToolName: m.Name(), Error: fmt.Errorf("%s: %s", err, stderr.String())}
	}

	fmt.Printf("=== MSF RAW OUTPUT ===\n%s\n======================\n", out.String())

	return Result{ToolName: m.Name(), Output: parseMsfOutput(out.String())}
}

func parseMsfOutput(raw string) string {
	var sb strings.Builder

	lines := strings.Split(raw, "\n")

	skip := regexp.MustCompile(`(?i)^(metasploit|msf6|msf5|msf >|msf6 >|=[=]+|\*=+|,=+|Type|----|encryption\.|Note|Value)`)

	interesting := regexp.MustCompile(`(?i)(` +
		`READ/WRITE` +     
	    `|SSH server version` +
		`|Key Fingerprint` +
		`|Server Information` +
		`|Scanned .* host` +
		`|Auxiliary module execution completed` +
		`|exploit completed` +
		`|Meterpreter session` +
		`|Command shell session` +
		`|error|fail` +
		`|\[\+\]` +
		`|\[\*\]` +
		`|\[!\]` +
		`)`)

	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if skip.MatchString(l) {
			continue
		}
		if interesting.MatchString(l) {
			sb.WriteString(l + "\n")
		}
	}

	if sb.Len() == 0 {
		return "Módulo ejecutado sin output relevante (el servicio puede no responder o no devolver banner)."
	}
	return sb.String()
}

func detectLocalIP() string {
	out := runCmd("hostname", "-I")
	fields := strings.Fields(out)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

// =====================
// MSF Module Map
// =====================

var (
	msfModules   map[string]string
	msfModulesMu sync.Once
)

const (
    msfExploitsPath = "/opt/metasploit-framework/embedded/framework/modules/exploits"
    msfCachePath    = "/root/.deyaclaw/msf_modules.json"
)

var msfCveRe = regexp.MustCompile(`'CVE',\s*'(\d{4}-\d+)'`)

func GetMSFModule(cveID string) (string, bool) {
	msfModulesMu.Do(func() {
		msfModules = loadMSFModules()
	})
	mod, ok := msfModules[cveID]
	return mod, ok
}

func loadMSFModules() map[string]string {
	if info, err := os.Stat(msfCachePath); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			if data, err := os.ReadFile(msfCachePath); err == nil {
				var m map[string]string
				if json.Unmarshal(data, &m) == nil {
					return m
				}
			}
		}
	}

	result := make(map[string]string)

	filepath.Walk(msfExploitsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".rb") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			for _, m := range msfCveRe.FindAllStringSubmatch(scanner.Text(), -1) {
				key := "CVE-" + m[1]
				if _, exists := result[key]; !exists {
					rel, _ := filepath.Rel("/opt/metasploit-framework/embedded/framework/modules", path)
					rel = strings.TrimSuffix(rel, ".rb")
					// Normalizar: "exploits/..." → "exploit/..."
					rel = strings.TrimPrefix(rel, "exploits/")
					rel = "exploit/" + rel
					result[key] = rel
					
				}
			}
		}
		return nil
	})

	if b, err := json.MarshalIndent(result, "", "  "); err == nil {
		os.MkdirAll(filepath.Dir(msfCachePath), 0755)
		os.WriteFile(msfCachePath, b, 0644)
	}
	return result
}
