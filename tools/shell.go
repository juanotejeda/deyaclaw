package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Whitelist de comandos permitidos
var allowedCommands = map[string]bool{
	"nmap":      true,
	"whois":     true,
	"dig":       true,
	"curl":      true,
	"ping":      true,
	"traceroute": true,
	"netstat":   true,
	"ss":        true,
	"gobuster":  true,
	"nikto":     true,
}

type ShellTool struct{}

func (s *ShellTool) Name() string { return "shell" }

func (s *ShellTool) Description() string {
	return "Ejecuta comandos del sistema. SOLO para: nmap, whois, dig, curl, ping, " +
		"traceroute, netstat, ss, gobuster, nikto. " +
		"NO usar para msfconsole ni Metasploit (usar tool 'metasploit' para eso)."
}

func (s *ShellTool) Execute(params map[string]string) Result {
	cmd := params["cmd"]
	if cmd == "" {
		return Result{ToolName: s.Name(), Error: fmt.Errorf("parámetro 'cmd' requerido")}
	}

	parts := strings.Fields(cmd)
	binary := parts[0]

	if !allowedCommands[binary] {
		return Result{
			ToolName: s.Name(),
			Error:    fmt.Errorf("comando '%s' no permitido. Whitelist: %v", binary, allowedCommands),
		}
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	c := exec.Command(parts[0], parts[1:]...)
	c.Stdout = &out
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		return Result{ToolName: s.Name(), Error: fmt.Errorf("%s: %s", err, stderr.String())}
	}

	return Result{ToolName: s.Name(), Output: out.String()}
}
