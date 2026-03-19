package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

type CheckEnvTool struct{}

func (c *CheckEnvTool) Name() string { return "checkenv" }

func (c *CheckEnvTool) Description() string {
	return "Verifica si el toolkit de pentest nivel 1 está instalado. Sin params."
}

func (c *CheckEnvTool) Execute(params map[string]string) Result {
	type entry struct {
		cmd  string
		pkg  string
	}

	toolkit := []entry{
		{"nmap", "nmap"},
		{"whatweb", "whatweb"},
		{"dig", "dnsutils"},
		{"whois", "whois"},
		{"gobuster", "gobuster"},
		{"nikto", "nikto"},
		{"hydra", "hydra"},
		{"theHarvester", "theharvester"},
		{"mitmproxy", "mitmproxy"},
		{"msfconsole", "metasploit-framework"},
	}

	var sb strings.Builder
	var missing []string

	sb.WriteString("=== PENTEST TOOLKIT — NIVEL 1 ===\n")

	for _, t := range toolkit {
		path, err := exec.LookPath(t.cmd)
		if err == nil {
			sb.WriteString(fmt.Sprintf("  ✅ %-20s %s\n", t.cmd, path))
		} else {
			sb.WriteString(fmt.Sprintf("  ❌ %-20s no encontrado\n", t.cmd))
			missing = append(missing, t.pkg)
		}
	}

	// SecLists
	seclistsPath := "/usr/share/seclists"
	out := runCmd("test", "-d", seclistsPath)
	if !strings.Contains(out, "error") {
		sb.WriteString(fmt.Sprintf("  ✅ %-20s %s\n", "seclists", seclistsPath))
	} else {
		sb.WriteString(fmt.Sprintf("  ❌ %-20s no encontrado\n", "seclists"))
		missing = append(missing, "seclists")
	}

	sb.WriteString("\n")
	if len(missing) == 0 {
		sb.WriteString("✅ Entorno completo. Listo para operar.\n")
	} else {
		sb.WriteString(fmt.Sprintf("⚠️  Faltan %d herramientas: %s\n", len(missing), strings.Join(missing, ", ")))
		sb.WriteString("💡 Corré: bash scripts/setup-pentest.sh\n")
	}

	return Result{ToolName: c.Name(), Output: sb.String()}
}
