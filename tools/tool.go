package tools

import "fmt"

// Result es el resultado de ejecutar una herramienta
type Result struct {
	ToolName string
	Output   string
	Error    error
}

func (r Result) String() string {
	if r.Error != nil {
		return fmt.Sprintf("[%s] ERROR: %s", r.ToolName, r.Error)
	}
	return fmt.Sprintf("[%s]\n%s", r.ToolName, r.Output)
}

// Tool es la interfaz que toda herramienta debe implementar
type Tool interface {
	Name() string
	Description() string
	Execute(params map[string]string) Result
}

// Registry contiene todas las herramientas disponibles
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) All() map[string]Tool {
	return r.tools
}

// ToolList devuelve descripción de todas las tools para el system prompt
func (r *Registry) ToolList() string {
	var sb string
	for _, t := range r.tools {
		sb += fmt.Sprintf("- %s: %s\n", t.Name(), t.Description())
	}
	return sb
}
