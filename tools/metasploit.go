package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type MetasploitTool struct{}

func (m *MetasploitTool) Name() string { return "metasploit" }

func (m *MetasploitTool) Description() string {
	return "Ejecuta módulos de Metasploit Framework usando msfconsole. " +
		"USA ESTA TOOL (no shell) para cualquier módulo de Metasploit. " +
		"Params: module (ej: auxiliary/scanner/ssh/ssh_version), " +
		"options (string: \"CLAVE=VALOR CLAVE2=VALOR2\", ej: \"RHOSTS=127.0.0.1 RPORT=22\"), " +
		"timeout (opcional, segundos, default 120)."
}

// Execute ejecuta msfconsole con un script generado a partir de los params.
// Espera SIEMPRE que params["options"] sea un string plano "K=V K2=V2".
func (m *MetasploitTool) Execute(params map[string]string) Result {
	module := params["module"]
	if module == "" {
		return Result{ToolName: m.Name(), Error: fmt.Errorf("parámetro 'module' requerido")}
	}

	// Timeout configurable
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

	// Auto-detectar LHOST si es exploit y no vino en options
	rawOptions := params["options"]
	if strings.HasPrefix(module, "exploit/") && !strings.Contains(rawOptions, "LHOST=") {
		if lhost := detectLocalIP(); lhost != "" {
			script.WriteString(fmt.Sprintf("set LHOST %s; ", lhost))
		}
	}

	// Parsear options string "K=V K2=V2"
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

	fmt.Printf("=== MSF SCRIPT ===\n%s\n==================; ", script.String())
	

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Result{ToolName: m.Name(), Error: fmt.Errorf("timeout: módulo superó %s", timeout)}
		}
		// Módulo no encontrado
		if strings.Contains(out.String(), "Failed to load module") ||
			strings.Contains(out.String(), "Invalid module") {
			return Result{ToolName: m.Name(), Error: fmt.Errorf("módulo no encontrado: %s", module)}
		}
		return Result{ToolName: m.Name(), Error: fmt.Errorf("%s: %s", err, stderr.String())}
	}

	fmt.Printf("=== MSF RAW OUTPUT ===\n%s\n======================; ", out.String()) // <-- acá

	return Result{ToolName: m.Name(), Output: out.String()}
}

// parseMsfOutput filtra el ruido de msfconsole y devuelve las líneas relevantes.
func parseMsfOutput(raw string) string {
	var sb strings.Builder

	lines := strings.Split(raw, "")

	skip := regexp.MustCompile(`(?i)^(metasploit|msf6|msf5|msf >|msf6 >|=[=]+|\*=+|,=+|Type|----|encryption\.|Note|Value)`)

	interesting := regexp.MustCompile(`(?i)(` +
		`SSH server version` +
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
			sb.WriteString(l + "")
		}
	}

	if sb.Len() == 0 {
		return "Módulo ejecutado sin output relevante (el servicio puede no responder o no devolver banner)."
	}
	return sb.String()
}


// detectLocalIP obtiene la IP local principal (primer IP reportada por hostname -I).
func detectLocalIP() string {
	out := runCmd("hostname", "-I")
	fields := strings.Fields(out)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}
