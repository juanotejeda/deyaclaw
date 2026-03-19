package tools

import "strings"

// patrones que podrían inyectar instrucciones al LLM
var injectionPatterns = []string{
	"ignore previous instructions",
	"ignore all previous",
	"disregard previous",
	"you are now",
	"new instructions:",
	"system prompt:",
	"[system]",
	"[user]",
	"[assistant]",
	"<system>",
	"<user>",
	"<assistant>",
	"forget everything",
	"olvidá todo",
	"ignorá las instrucciones",
	"nuevas instrucciones:",
	"sos ahora",
}

func SanitizeOutput(output string) string {
	lower := strings.ToLower(output)
	for _, pattern := range injectionPatterns {
		if strings.Contains(lower, pattern) {
			// Reemplazar línea completa que contiene el patrón
			lines := strings.Split(output, "\n")
			var clean []string
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), pattern) {
					clean = append(clean, "[línea removida por seguridad]")
				} else {
					clean = append(clean, line)
				}
			}
			output = strings.Join(clean, "\n")
		}
	}
	return output
}
