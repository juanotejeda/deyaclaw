package tools

import (
	"bytes"
	"fmt"
	"os/exec"
)

type WhoisTool struct{}

func (w *WhoisTool) Name() string { return "whois" }

func (w *WhoisTool) Description() string {
	return "Consulta información de registro de un dominio o IP. Params: target (dominio o IP)"
}

func (w *WhoisTool) Execute(params map[string]string) Result {
	target := params["target"]
	if target == "" {
		return Result{ToolName: w.Name(), Error: fmt.Errorf("parámetro 'target' requerido")}
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("whois", target)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return Result{ToolName: w.Name(), Error: fmt.Errorf("%s: %s", err, stderr.String())}
	}

	return Result{ToolName: w.Name(), Output: out.String()}
}
