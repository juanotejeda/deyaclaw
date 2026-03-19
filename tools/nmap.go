package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type NmapTool struct{}

func (n *NmapTool) Name() string { return "nmap" }

func (n *NmapTool) Description() string {
	return "Escanea hosts y puertos. Params: target (IP o rango CIDR), flags (opcional, default: -sV -T4 --open)"
}

func (n *NmapTool) Execute(params map[string]string) Result {
	target := params["target"]
	if target == "" {
		return Result{ToolName: n.Name(), Error: fmt.Errorf("parámetro 'target' requerido")}
	}

	flags := params["flags"]
	if flags == "" {
		flags = "-sV -T4 --open"
	}

	args := strings.Fields(flags)
	args = append(args, "-oG", "-") // output parseable
	args = append(args, target)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "nmap", args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Result{ToolName: n.Name(), Error: fmt.Errorf("timeout: escaneo superó 120s")}
		}
		return Result{ToolName: n.Name(), Error: fmt.Errorf("%s: %s", err, stderr.String())}
	}

	parsed := parseNmapGrep(out.String())
	return Result{ToolName: n.Name(), Output: parsed}
}

// parseNmapGrep convierte output -oG en texto estructurado para el LLM
func parseNmapGrep(raw string) string {
	var sb strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "Host:") {
			parts := strings.Split(line, "\t")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if strings.HasPrefix(p, "Host:") || strings.HasPrefix(p, "Ports:") || strings.HasPrefix(p, "Status:") {
					sb.WriteString(p + "\n")
				}
			}
		}
	}
	if sb.Len() == 0 {
		return "Sin hosts activos encontrados."
	}
	return sb.String()
}
