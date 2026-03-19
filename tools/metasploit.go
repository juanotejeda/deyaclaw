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
	return "Ejecuta módulos de Metasploit. Params: module (ej: auxiliary/scanner/ssh/ssh_version), " +
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
	script.WriteString(fmt.Sprintf("use %s\n", module))

	// Auto-detectar LHOST si es exploit y no vino en options
	rawOptions := params["options"]
	if strings.HasPrefix(module, "exploit/") && !strings.Contains(rawOptions, "LHOST=") {
		if lhost := detectLocalIP(); lhost != "" {
			script.WriteString(fmt.Sprintf("set LHOST %s\n", lhost))
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
					script.WriteString(fmt.Sprintf("set %s %s\n", key, val))
				}
			}
		}
	}

	script.WriteString("run\n")
	script.WriteString("exit\n")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "msfconsole", "-q", "-x", script.String())
	cmd.Stdout = &out
	cmd.Stderr = &stderr

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

	parsed := parseMsfOutput(out.String())
	return Result{ToolName: m.Name(), Output: parsed}
}

// parseMsfOutput intenta filtrar ruido de msfconsole y dejar solo lo relevante.
func parseMsfOutput(raw string) string {
	var sb strings.Builder

	lines := strings.Split(raw, "\n")

	// 1) Si encontramos explícitamente la versión SSH, la devolvemos clara.
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		// Ejemplo: "[*] 127.0.0.1 - SSH server version: SSH-2.0-OpenSSH_9.2p1 Debian-2+deb12u7"
		if strings.Contains(l, "SSH server version:") {
			sb.WriteString("[+] Resultado Metasploit — SSH version\n")
			sb.WriteString(l + "\n")
			return sb.String()
		}
	}

	// 2) Si no hay línea de versión explícita, extraemos todas las líneas interesantes.
	skip := regexp.MustCompile(`(?i)^(metasploit|msf6|msf >|=[=]+)$`)
	interesting := regexp.MustCompile(`(?i)(SSH server version|Key Fingerprint|Server Information|Scanned .* hosts|Auxiliary module execution completed|error|fail|\[\+\]|\[!\])`)

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
		return "Módulo ejecutado, pero no se encontró output relevante (puede que no haya hosts que respondan o que el servicio no devuelva banner)."
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
