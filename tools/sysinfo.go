package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type SysInfoTool struct{}

func (s *SysInfoTool) Name() string { return "sysinfo" }

func (s *SysInfoTool) Description() string {
	return "Obtiene información del sistema local: IPs, interfaces, OS, kernel y herramientas disponibles. Sin params."
}

func (s *SysInfoTool) Execute(params map[string]string) Result {
	var sb strings.Builder

	// IP e interfaces
	sb.WriteString("=== INTERFACES Y IPs ===\n")
	sb.WriteString(runCmd("ip", "addr", "show"))

	// Tabla de ruteo
	sb.WriteString("\n=== TABLA DE RUTEO ===\n")
	sb.WriteString(runCmd("ip", "route"))

	// OS y kernel
	sb.WriteString("\n=== SISTEMA OPERATIVO ===\n")
	sb.WriteString(runCmd("uname", "-a"))

	// Herramientas disponibles
	sb.WriteString("\n=== HERRAMIENTAS DISPONIBLES ===\n")
	toolsList := []string{
		"nmap", "msfconsole", "gobuster", "nikto",
		"curl", "dig", "netcat", "nc", "hydra",
		"sqlmap", "burpsuite", "wireshark", "tcpdump",
		"metasploit", "john", "hashcat",
	}
	for _, t := range toolsList {
		path := runCmd("which", t)
		path = strings.TrimSpace(path)
		if path != "" && !strings.Contains(path, "not found") {
			sb.WriteString(fmt.Sprintf("  ✅ %-15s %s\n", t, path))
		}
	}

	// Puertos locales en escucha
	sb.WriteString("\n=== PUERTOS EN ESCUCHA ===\n")
	sb.WriteString(runCmd("ss", "-tlnp"))

	return Result{ToolName: s.Name(), Output: sb.String()}
}

func runCmd(name string, args ...string) string {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("error: %s\n", stderr.String())
	}
	return out.String()
}
