package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

type PortScanTool struct{}

func (t *PortScanTool) Name() string { return "portscan" }
func (t *PortScanTool) Description() string {
	return "Escanea puertos de un host usando nmap. Params: target (IP o dominio), ports (ej: '1-1000', '80,443,8080', 'top100'), mode (quick|full|stealth)"
}
func (t *PortScanTool) Params() []string { return []string{"target", "ports", "mode"} }

func (t *PortScanTool) Execute(params map[string]string) Result {
	target := params["target"]
	if target == "" {
		return Result{
			ToolName: t.Name(),
			Error:    fmt.Errorf("parámetro 'target' requerido"),
		}
	}

	ports := params["ports"]
	mode := params["mode"]
	if mode == "" {
		mode = "quick"
	}

	args, description := buildNmapArgs(target, ports, mode)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Escaneando %s [modo: %s]\n", target, description))
	sb.WriteString(fmt.Sprintf("Comando: nmap %s\n\n", strings.Join(args, " ")))

	cmd := exec.Command("nmap", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{
			ToolName: t.Name(),
			Error:    fmt.Errorf("error ejecutando nmap: %w\n%s", err, string(out)),
		}
	}

	sb.WriteString(string(out))

	return Result{
		ToolName: t.Name(),
		Output:   sb.String(),
	}
}

func buildNmapArgs(target, ports, mode string) ([]string, string) {
	var args []string
	var description string

	switch mode {
	case "full":
		// Escaneo completo con detección de versión y OS
		args = []string{"-sV", "--open", "-T4"}
		description = "completo (versión + OS)"
	case "stealth":
		// SYN scan silencioso
		args = []string{"-sS", "-T2", "--open"}
		description = "stealth (SYN silencioso)"
	default:
		// Quick: top ports con detección básica
		args = []string{"-sV", "--open", "-T4"}
		description = "rápido (top ports + versión)"
	}

	// Puertos
	switch ports {
	case "", "top100":
		args = append(args, "--top-ports", "100")
	case "top1000":
		args = append(args, "--top-ports", "1000")
	case "all":
		args = append(args, "-p-")
	default:
		args = append(args, "-p", ports)
	}

	// Output legible + scripts básicos
	args = append(args, "-sC", "--reason", target)

	return args, description
}
