package executor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/juano/deyaclaw/tools"
)

type Action struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Reason string                 `json:"reason"`
}

type Executor struct {
	registry   *tools.Registry
	supervised bool
}

func New(registry *tools.Registry, supervised bool) *Executor {
	return &Executor{
		registry:   registry,
		supervised: supervised,
	}
}

func ParseAction(text string) (*Action, bool) {
	start := strings.Index(text, "{")
	end   := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, false
	}

	jsonStr := text[start : end+1]

	var action Action
	if err := json.Unmarshal([]byte(jsonStr), &action); err != nil {
		return nil, false
	}

	if action.Action == "" {
		return nil, false
	}

	return &action, true
}

func normalizeParams(raw map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range raw {
		switch val := v.(type) {
		case string:
			result[k] = val
		case []interface{}:
			parts := []string{}
			for _, item := range val {
				if s, ok := item.(string); ok {
					parts = append(parts, s)
				}
			}
			result[k] = strings.Join(parts, " ")
		default:
			result[k] = fmt.Sprintf("%v", val)
		}
	}
	return result
}

// Run ejecuta una action, pidiendo confirmación si está en modo supervisado
func (e *Executor) Run(action *Action, confirm func(string) bool) tools.Result {
	tool, ok := e.registry.Get(action.Action)
	if !ok {
		return tools.Result{
			ToolName: action.Action,
			Error:    fmt.Errorf("herramienta '%s' no disponible", action.Action),
		}
	}

	params := normalizeParams(action.Params)

	if e.supervised {
		cmd := buildCommandPreview(action.Action, params)
		if !confirm(cmd) {
			return tools.Result{
				ToolName: action.Action,
				Error:    fmt.Errorf("acción cancelada por el usuario"),
			}
		}
	}

	return tool.Execute(params)
}

// RunDirect ejecuta sin pedir confirmación (ya fue aprobada externamente)
func (e *Executor) RunDirect(action *Action) tools.Result {
	tool, ok := e.registry.Get(action.Action)
	if !ok {
		return tools.Result{
			ToolName: action.Action,
			Error:    fmt.Errorf("herramienta '%s' no disponible", action.Action),
		}
	}
	return tool.Execute(normalizeParams(action.Params))
}

func buildCommandPreview(name string, params map[string]string) string {
	var parts []string
	parts = append(parts, name)
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, " ")
}
