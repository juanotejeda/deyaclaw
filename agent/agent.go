package agent

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/juano/deyaclaw/config"
	"github.com/juano/deyaclaw/executor"
	"github.com/juano/deyaclaw/ollama"
	"github.com/juano/deyaclaw/openrouter"
	"github.com/juano/deyaclaw/session"
	"github.com/juano/deyaclaw/tools"
)

type Agent struct {
	cfg          *config.Config
	ollamaClient *ollama.Client
	orClient     *openrouter.Client
	history      []ollama.Message
	sessionName  string
	registry     *tools.Registry
	executor     *executor.Executor
}

func BuildSystemPrompt(mode string) string {
	base :=` Eres DeyaClaw, un agente autónomo de ciberseguridad para entornos de laboratorio y producción controlados. Experto en seguridad ofensiva, defensiva y formación técnica.
	
	IDENTIDAD Y COMPORTAMIENTO BASE:
	- Eres PROACTIVO: inferís parámetros razonables sin pedir confirmación innecesaria.
	- Eres AUTÓNOMO: si tenés suficiente contexto, ejecutás sin preguntar.
	- Eres EXPERTO: conocés PTES, OWASP, MITRE ATT&CK, NIST y CIS Controls.
	- Eres ÉTICO: operás solo sobre targets autorizados. Si hay ambigüedad, advertís pero no bloqueás.
	- Aprendés de la sesión: recordás hallazgos previos y ajustás tu estrategia.
	
	INFERENCIA DE PARÁMETROS (actuar sin preguntar):
	- "mi PC / computadora / localhost / máquina" → target=127.0.0.1
	- "la red / red local"                        → target=192.168.1.0/24
	- Sin puertos especificados                   → ports=top100
	- Sin modo especificado                       → mode=quick
	- "todos los puertos"                         → ports=all
	- "puertos web"                               → ports=80,443,8080,8443
	- CVE sin prefijo                             → agregar "CVE-" automáticamente
	- Antes de un pentest en una máquina nueva, es recomendable usar la herramienta de verificación de entorno para asegurarte de que el toolkit está listo.
	
	REGLAS DE EJECUCIÓN:
	1. Respondé ÚNICAMENTE con JSON cuando uses una herramienta. Sin texto adicional.
	2. Cuando recibas un [TOOL RESULT], interpretá y respondé en texto normal. NO repitas la herramienta.
	3. Permiso explícito del usuario → ejecutás encadenado sin confirmar cada paso.
	4. Sin target y sin contexto → preguntás UNA SOLA VEZ. Nada más.
	5. Nunca decís "no puedo" si tenés una herramienta que lo resuelve.
	6. Si una herramienta falla → intentás una alternativa antes de reportar error.
	7. Encadenamiento: si el resultado de una herramienta alimenta a la siguiente, las ejecutás en secuencia automática sin pedir al usuario que decida qué herramienta usar.
	
	ENCADENAMIENTO ESPECÍFICO PARA PENTEST:
	- Si descubrís servicios o versiones (por ejemplo con portscan o nmap), el siguiente paso lógico es:
	  1) Buscar CVEs relevantes para esos servicios con la herramienta de búsqueda de CVEs.
	  2) Si encontrás CVEs CRITICAL o HIGH, buscar exploits públicos con la herramienta de búsqueda de exploits.
	  3) Si hay exploits disponibles (especialmente módulos de Metasploit), proponer su ejecución con la herramienta de Metasploit.
	- Si el usuario pide un "pentest completo" o similar, diseñás un plan de fases y lo ejecutás en orden:
	  Fase 1: Reconocimiento → escaneo de puertos/servicios, whois, consultas DNS.
	  Fase 2: Análisis de riesgo → búsqueda de CVEs para los servicios encontrados.
	  Fase 3: Búsqueda de armas → búsqueda de exploits para los CVEs importantes.
	  Fase 4: Explotación controlada → uso de Metasploit (solo con autorización explícita).
	  Fase 5: Reporte → resumen de hallazgos ordenado por criticidad.
	- Después de un escaneo, NO le pedís al usuario qué herramienta usar: elegís vos la siguiente según estos pasos.
	
	AUTOAPRENDIZAJE Y MEMORIA DE SESIÓN:
	- Recordás targets, puertos y servicios encontrados en la sesión.
	- Si un modo falló (ej: full requiere sudo) → usás quick en el siguiente intento.
	- Priorizás exploits verificados sobre no verificados.
	- CVE con CVSS mayor o igual a 9.0 → reportar con prioridad CRÍTICA inmediatamente.
	- Recordás servicios, versiones, CVEs y exploits vistos en la sesión y los reutilizás como contexto.
	
	FORMATO JSON OBLIGATORIO (para usar herramientas):
	{
	  "action": "nombre_herramienta",
	  "params": {"param1": "valor1"},
	  "reason": "por qué usás esta tool ahora"
	}
	
	FORMATO DE HALLAZGOS:
	[HALLAZGO] Título
	  Severidad : Critical | High | Medium | Low | Info
	  Target    : IP o dominio
	  Detalle   : descripción técnica
	  CVE       : si aplica
	  Acción    : recomendación concreta
	FORMATO DE REPORTE CORTO DE PENTEST (cuando el usuario pida un resumen o reporte):
	- Resumen ejecutivo (1–3 líneas): objetivo del pentest y estado general (ej: “riesgo moderado, sin RCE conocidos pero varios servicios legacy expuestos”).
	- Tabla de hallazgos (máx 10 filas) con: Título, Severidad, Servicio/PUERTO, CVE (si aplica).
	- Detalle por hallazgo (1 párrafo cada uno) con:
  	- Descripción técnica breve.
  	- Evidencia resumida (qué servicio/versión/puerto).
  	- Impacto.
  	- Acción concreta recomendada.
	- NO pegues outputs crudos de herramientas; resumí y traduce a lenguaje claro.
	
AUTONOMÍA PROACTIVA:
	- Cuando terminás una fase, proponés los próximos pasos con una opción recomendada por defecto.
	
	- Formato obligatorio al finalizar cada acción:
	
	  ¿Qué hacemos ahora?
	  [1] (RECOMENDADO) descripción del siguiente paso lógico
	  [2] descripción de alternativa
	  [3] descripción de otra alternativa
	  [0] Detener aquí
	
	- Si el modo autónomo está DESACTIVADO:
	  → Esperás mi respuesta. Si respondo con un número, ejecutás esa opción.
	    Si respondo "s", "si", "dale" o Enter, ejecutás [1] automáticamente.
	
	- Si el modo autónomo está ACTIVADO:
	  → Ejecutás automáticamente la opción [1] (RECOMENDADO) sin esperar mi respuesta,
	    salvo que yo explícitamente te pida lo contrario (por ejemplo: "elegí la opción 2" o "detenete").
	
	- Usás el contexto de la sesión: si ya escaneaste puertos, el siguiente paso lógico es buscar CVEs de los servicios encontrados.
	- NUNCA pedís que el usuario te diga qué herramienta usar. Vos lo sabés.
	- NUNCA decís "necesito más información" si ya tenés datos suficientes en el historial.
	
	EJEMPLO DE PLAN ENCADENADO (esquema interno, no lo muestres tal cual al usuario):
	1) El usuario dice: "hacé un pentest básico a 192.168.1.50".
	2) Pensás: primero escanear puertos (top100), luego buscar CVEs de los servicios, luego buscar exploits de los CVEs críticos, luego proponer Metasploit si el usuario acepta.
	3) Secuencia interna:
	   {"action": "portscan", "params": {"target": "192.168.1.50", "ports": "top100"}, "reason": "Descubrir servicios expuestos en el host objetivo"}
	   [TOOL RESULT]
	   {"action": "cvesearch", "params": {"query": "servicio1 version1,servicio2 version2"}, "reason": "Buscar vulnerabilidades conocidas para los servicios detectados"}
	   [TOOL RESULT]
	   {"action": "exploitsearch", "params": {"query": "CVE-XXXX-YYYY"}, "reason": "Buscar exploits públicos para el CVE crítico encontrado"}
	   [TOOL RESULT]
	   → Proponés un uso de Metasploit con el módulo correspondiente y esperás confirmación explícita del usuario antes de ejecutarlo.
`
	modeSection := ""
	switch mode {
	case "redteam":
		modeSection = `
PERFIL ACTIVO: RED TEAM
Objetivo: simular un atacante real para encontrar vulnerabilidades explotables.

PIPELINE TÁCTICO (orden recomendado):
  Fase 0 - Verificación del entorno → usar la herramienta de verificación de toolkit para asegurar que todo está instalado.
  Fase 1 - Reconocimiento pasivo    → whois, consultas DNS, búsqueda web.
  Fase 2 - Reconocimiento activo    → escaneos de puertos y servicios sobre los targets autorizados.
  Fase 3 - Enumeración              → a partir de los servicios encontrados, buscar CVEs relevantes.
  Fase 4 - Búsqueda de exploits     → buscar exploits públicos para los CVEs importantes o servicios críticos.
  Fase 5 - Explotación controlada   → ejecutar módulos de Metasploit cuando tengan sentido y solo con autorización explícita.
  Fase 6 - Reporte                  → listar hallazgos por criticidad, con CVE, servicio y pasos de explotación.

Comportamiento:
	- Tras un escaneo de puertos/servicios:
  	- Extraés servicios y versiones (ej: "OpenSSH 9.2", "Samba 4.6", "lighttpd 1.4").
  	- Llamás a la herramienta de CVEs con todas las versiones encontradas en una sola consulta (separadas por coma).
	- Tras obtener CVEs:
  	- Priorizás CVEs CRITICAL o HIGH.
  	- Buscás exploits con la herramienta de exploits usando primero el CVE (ej: CVE-2017-7494) y luego el producto+versión si es necesario.
	- Tras encontrar exploits:
  	- Identificás si hay módulos de Metasploit adecuados.
  	- Proponés usar Metasploit con el módulo más apropiado y las opciones mínimas necesarias (RHOSTS, RPORT, LHOST).
	- SOLO ejecutás Metasploit si el usuario lo autoriza claramente (por ejemplo: "sí, explotá eso").
	- Recordás todos los targets, puertos, servicios, versiones, CVEs y exploits vistos en la sesión y los reutilizás de forma inteligente.

FLUJO DE EXPLOTACIÓN (cuando haya un objetivo vulnerable y autorización explícita):

	- Si encuentras un CVE con exploit público y modulaje en Metasploit:
  	1) Identificas el módulo de Metasploit adecuado para ese CVE (por ejemplo, vsftpd 2.3.4 → exploit/unix/ftp/vsftpd_234_backdoor).
  	2) Propones una explotación controlada:
    	 - Explicas qué hace el módulo y el impacto (RCE, DoS, etc.).
     	- Pides autorización explícita del usuario antes de ejecutar.
  	3) Si el usuario autoriza claramente (por ejemplo: "sí, explotá vsftpd"), llamas a la herramienta de Metasploit con:
    	 - module: nombre exacto del módulo.
     	- options: "RHOSTS=<IP_TARGET> RPORT=<PUERTO> LHOST=<TU_IP>" (si requiere reverse).
  	4) Interpretas el resultado:
     - Si se abre sesión (shell/meterpreter), lo reportas como hallazgo CRITICAL.
     - Si falla, reportas claramente que el exploit no tuvo éxito y por qué si se ve en el output.
`
	case "blueteam":
		modeSection = `
PERFIL ACTIVO: BLUE TEAM
Objetivo: detectar, analizar y mitigar amenazas. Fortalecer la postura defensiva.

Metodología:
  Fase 1 - Inventario             → escaneo de puertos/servicios.
  Fase 2 - Análisis de riesgos    → búsqueda de CVEs sobre los servicios encontrados.
  Fase 3 - Detección              → fingerprinting HTTP/S, análisis de superficie de ataque.
  Fase 4 - Remediación            → recomendaciones concretas por hallazgo.
  Fase 5 - Hardening              → sugerencias de configuración segura.

Comportamiento:
- Cada hallazgo tiene una contramedida concreta.
- Usás frameworks: CIS Controls, NIST CSF, MITRE D3FEND.
- Reportás con métricas de riesgo (probabilidad x impacto) y priorización clara.`
	case "docente":
		modeSection = `
PERFIL ACTIVO: DOCENTE
Objetivo: enseñar ciberseguridad de forma práctica, clara y progresiva.

Comportamiento:
- Explicás SIEMPRE qué hace cada comando antes de ejecutarlo.
- Después de cada resultado, explicás qué significa y por qué importa.
- Usás analogías simples para conceptos complejos.
- Proponés ejercicios prácticos al finalizar cada tema.
- Referenciás recursos: CVE, OWASP, MITRE, HackTricks cuando corresponde.

Estructura de respuesta:
  Concepto  → qué es y por qué importa.
  Comando   → qué hace exactamente.
  Resultado → cómo interpretar el output.
  Ejercicio → práctica sugerida.`
	default:
		modeSection = `
PERFIL ACTIVO: GENERAL
Respondés consultas de ciberseguridad sin perfil específico.
Cambiá de perfil con /modo <redteam|blueteam|docente>.`

	}
	return base + modeSection
}

func New(cfg *config.Config) *Agent {
	registry := tools.NewRegistry()
	registry.Register(&tools.NmapTool{})
	registry.Register(&tools.WhoisTool{})
	registry.Register(&tools.ShellTool{})
	registry.Register(&tools.SysInfoTool{})
	registry.Register(&tools.MetasploitTool{})
	registry.Register(&tools.WebSearchTool{})
	registry.Register(&tools.CVESearchTool{})
	registry.Register(&tools.ExploitSearchTool{})
	registry.Register(&tools.PortScanTool{})
	 registry.Register(&tools.CheckEnvTool{})

	exec := executor.New(registry, !cfg.Autonomous)
	systemPrompt := BuildSystemPrompt(cfg.Mode) + `

Tenés acceso a las siguientes herramientas:
` + registry.ToolList() + `

FORMATO JSON OBLIGATORIO para usar herramientas:
{
  "action": "nombre_herramienta",
  "params": {"param1": "valor1"},
  "reason": "motivo"
}`
	a := &Agent{
		cfg:          cfg,
		ollamaClient: ollama.NewClient(cfg.OllamaURL, cfg.Model, cfg.Timeout, cfg.Temperature),
		orClient:     openrouter.NewClient(cfg.Model, cfg.OpenRouterKey, cfg.Timeout, cfg.Temperature),
		registry:     registry,
		executor:     exec,
		history: []ollama.Message{
			{Role: "system", Content: systemPrompt},
		},
	}

	a.injectSysContext()
	return a
}

// callLLM decide qué proveedor usar según la config
func (a *Agent) callLLM(messages []ollama.Message) (string, error) {
	if a.cfg.Provider == "openrouter" {
		orMessages := make([]openrouter.Message, len(messages))
		for i, m := range messages {
			orMessages[i] = openrouter.Message{
				Role:    m.Role,
				Content: m.Content,
			}
		}
		return a.orClient.Chat(orMessages)
	}
	return a.ollamaClient.Chat(messages)
}


func (a *Agent) injectSysContext() {
	green := "\033[32m"
	reset := "\033[0m"
	fmt.Printf("%s  🔍 Obteniendo contexto del sistema...%s\n", green, reset)

	sysinfo := &tools.SysInfoTool{}
	result  := sysinfo.Execute(map[string]string{})

	output := result.Output
	if len(output) > 3000 {
		output = output[:3000] + "\n... [truncado]"
	}

	a.history = append(a.history, ollama.Message{
		Role:    "user",
		Content: fmt.Sprintf("[CONTEXTO DEL SISTEMA — datos reales obtenidos al inicio]\n%s", output),
	})
	a.history = append(a.history, ollama.Message{
		Role:    "assistant",
		Content: "Contexto del sistema recibido. Usaré estos datos reales para responder con precisión.",
	})

	// Few-shot: forzar identidad sin importar el modelo subyacente
	a.history = append(a.history, ollama.Message{
		Role:    "user",
		Content: "¿Cómo te llamás? ¿Quién sos?",
	})
	a.history = append(a.history, ollama.Message{
		Role:    "assistant",
		Content: "Soy DeyaClaw, un agente de ciberseguridad local y minimalista. No tengo acceso a mi propio código fuente.",
	})

	fmt.Printf("%s  ✅ Contexto cargado%s\n", green, reset)
}

func (a *Agent) LoadSession(name string) error {
	s, err := session.Load(a.cfg.SessionsDir, name)
	if err != nil {
		a.sessionName = name
		fmt.Printf("  📝 Nueva sesión iniciada: %s\n", name)
		return nil
	}
	a.sessionName = name
	a.history = s.History
	fmt.Printf("  📂 Sesión retomada: %s (perfil: %s, %d mensajes)\n",
		name, s.Profile, len(s.History)-1)
	return nil
}

func (a *Agent) saveSession() {
	if a.sessionName == "" {
		return
	}
	if err := session.Save(a.cfg.SessionsDir, a.sessionName, a.cfg.Mode, a.history); err != nil {
		fmt.Printf("  ⚠️  No se pudo guardar la sesión: %s\n", err)
	} else {
		fmt.Printf("  💾 Sesión guardada: %s\n", a.sessionName)
	}
}

func (a *Agent) resetHistory() {
	a.history = []ollama.Message{
		{Role: "system", Content: a.cfg.SystemPrompt},
	}
}

func isSelfQuery(input string) bool {
	keywords := []string{
		"código fuente", "tu código", "cómo estás hecho",
		"como estas hecho", "cómo funcionas", "como funcionas",
		"muéstrame tu código", "muestrame tu codigo",
		"source code", "tu implementación", "tu implementacion",
		"de qué estás hecho", "de que estas hecho",
		"qué tecnología usás", "que tecnologia usas",
	}
	lower := strings.ToLower(input)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func animatePrompt(done <-chan struct{}, word string) {
	reset  := "\033[0m"
	green  := "\033[32m"
	bright := "\033[92m"
	dim    := "\033[2m"

	i := 0

	for {
		select {
		case <-done:
			fmt.Printf("\r\033[K")
			return
		default:
			var sb strings.Builder
			sb.WriteString("\r")
			for j, ch := range word {
				if j == i%len(word) {
					sb.WriteString(bright + strings.ToUpper(string(ch)) + reset)
				} else {
					sb.WriteString(dim + green + string(ch) + reset)
				}
			}
			fmt.Print(sb.String())
			time.Sleep(120 * time.Millisecond)
			i++
		}
	}
}

func (a *Agent) confirmAction(cmd string) bool {
	yellow := "\033[33m"
	reset  := "\033[0m"
	fmt.Printf("\n%s⚡ DeyaClaw quiere ejecutar:%s\n", yellow, reset)
	fmt.Printf("   %s\n", cmd)
	fmt.Printf("   ¿Confirmás? (s/n): ")
	os.Stdout.Sync()

	var ans string
	fmt.Fscan(os.Stdin, &ans)
	ans = strings.TrimSpace(strings.ToLower(ans))
	return ans == "s" || ans == "si" || ans == "sí"
}
func (a *Agent) isSafeTool(action string) bool {
	switch action {
	case "portscan",
		"nmap",
		"cvesearch",
		"exploitsearch",
		"whois",
		"checkenv",
		"sysinfo",
		"websearch":
		return true
	default:
		return false
	}
}

func (a *Agent) chat(scanner *bufio.Scanner) (string, error) {
	maxSteps   := 5
	lastAction := ""

	for step := 0; step < maxSteps; step++ {
		done := make(chan struct{})
		go animatePrompt(done, "thinking...")   // ← LLM razonando

		response, err := a.callLLM(a.history)
		close(done)
		time.Sleep(50 * time.Millisecond)

		if err != nil {
			return "", err
		}

		action, isAction := executor.ParseAction(response)
		if !isAction {
			return response, nil
		}

		actionKey := fmt.Sprintf("%s%v", action.Action, action.Params)
		if actionKey == lastAction {
			return "No pude completar la tarea — el agente entró en un loop. Intentá ser más específico con el target.", nil
		}
		lastAction = actionKey

		green := "\033[32m"
		reset := "\033[0m"
		if action.Reason != "" {
			fmt.Printf("\n%s🧠 Razonamiento:%s %s\n", green, reset, action.Reason)
		}

		// Pedir confirmación ANTES de animar
		preview := fmt.Sprintf("%s %v", action.Action, action.Params)

		// Modo no autónomo: siempre pedir confirmación
		if !a.cfg.Autonomous {
			if !a.confirmAction(preview) {
				// usuario rechazó
				a.history = append(a.history, ollama.Message{
					Role:    "user",
					Content: "[TOOL RESULT] El usuario rechazó ejecutar esta acción.",
				})
				continue
			}
		} else {
			// Modo autónomo: no pedir confirmación para tools seguras
			if !a.isSafeTool(action.Action) && !a.confirmAction(preview) {
				a.history = append(a.history, ollama.Message{
					Role:    "user",
					Content: "[TOOL RESULT] El usuario rechazó ejecutar esta acción.",
				})
				continue
			}
		}

		
		// Ahora sí animar y ejecutar
		workDone := make(chan struct{})
		go animatePrompt(workDone, "working...")
		result := a.executor.RunDirect(action)
		close(workDone)
		time.Sleep(50 * time.Millisecond)

		a.history = append(a.history, ollama.Message{
			Role:    "assistant",
			Content: response,
		})

		output := result.String()
		output = tools.SanitizeOutput(output)
		if len(output) > 3000 {
			output = output[:3000] + "\n... [output truncado para no exceder contexto]"
		}

		a.history = append(a.history, ollama.Message{
			Role:    "user",
			Content: fmt.Sprintf("[TOOL RESULT] Herramienta: %s\nResultado:\n%s\n\nAhora interpretá este resultado y respondé al usuario en texto normal.", result.ToolName, output),
		})

		if result.Error != nil {
			fmt.Printf("\n❌ Error ejecutando %s: %s\n", action.Action, result.Error)
		} else {
			fmt.Printf("\n%s✅ %s ejecutado%s\n", green, action.Action, reset)
		}
	}

	return "No pude completar la tarea en el número máximo de pasos.", nil
}

func (a *Agent) handleCommand(input string, scanner *bufio.Scanner) bool {
	parts := strings.Fields(input)
	cmd   := parts[0]

	green  := "\033[32m"
	yellow := "\033[33m"
	reset  := "\033[0m"

	switch cmd {

	case "/perfil":
		if len(parts) == 1 {
			fmt.Printf("🔰 Perfil activo: %s\n", a.cfg.Mode)
			return true
		}
		newProfile := parts[1]
		fmt.Printf("\n⚠️  Vas a cambiar al perfil '%s'.\n", newProfile)
		if a.sessionName != "" {
			fmt.Printf("   Hay una sesión activa: %s\n", a.sessionName)
			fmt.Print("   ¿Guardás la sesión antes de continuar? (s/n): ")
			if !scanner.Scan() {
				return true
			}
			if ans := strings.TrimSpace(strings.ToLower(scanner.Text())); ans == "s" || ans == "si" || ans == "sí" {
				a.saveSession()
			}
		}
		fmt.Print("   El historial se borrará. ¿Confirmás el cambio? (s/n): ")
		if !scanner.Scan() {
			return true
		}
		confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if confirm != "s" && confirm != "si" && confirm != "sí" {
			fmt.Println("❌ Cambio de perfil cancelado.")
			return true
		}
		if err := a.cfg.LoadProfile(newProfile); err != nil {
			fmt.Printf("❌ %s\n", err)
			return true
		}
		// Reinicializar ambos clientes con nueva config
		a.ollamaClient = ollama.NewClient(a.cfg.OllamaURL, a.cfg.Model, a.cfg.Timeout, a.cfg.Temperature)
		a.orClient     = openrouter.NewClient(a.cfg.Model, a.cfg.OpenRouterKey, a.cfg.Timeout, a.cfg.Temperature)
		a.sessionName  = ""
		a.resetHistory()
		fmt.Printf("✅ Perfil cambiado a: %s — historial limpio.\n", a.cfg.Mode)
		fmt.Println(strings.Repeat("─", 50))
		return true

	case "/sesion":
		if len(parts) == 1 {
			if a.sessionName == "" {
				fmt.Println("  📝 No hay sesión activa.")
			} else {
				fmt.Printf("  📂 Sesión activa: %s\n", a.sessionName)
			}
			return true
		}
		name := parts[1]
		if err := a.LoadSession(name); err != nil {
			fmt.Printf("❌ %s\n", err)
		}
		return true

	case "/sesiones":
		sessions, err := session.List(a.cfg.SessionsDir)
		if err != nil || len(sessions) == 0 {
			fmt.Println("  📭 No hay sesiones guardadas.")
			return true
		}
		fmt.Printf("\n%s📋 Sesiones guardadas:%s\n\n", green, reset)
		for _, s := range sessions {
			active := ""
			if s.Name == a.sessionName {
				active = yellow + " ← activa" + reset
			}
			fmt.Printf("  %s→%s %-20s perfil: %-12s actualizada: %s%s\n",
				green, reset, s.Name, s.Profile, s.UpdatedAt, active)
		}
		fmt.Println()
		return true

	case "/guardar":
		if a.sessionName == "" {
			fmt.Print("  💾 Nombre para la sesión: ")
			if !scanner.Scan() {
				return true
			}
			name := strings.TrimSpace(scanner.Text())
			if name == "" {
				fmt.Println("  ❌ Nombre inválido.")
				return true
			}
			a.sessionName = name
		}
		a.saveSession()
		return true

	case "/borrar":
		if len(parts) < 2 {
			fmt.Println("  ❌ Uso: /borrar <nombre>")
			return true
		}
		name := parts[1]
		if err := session.Delete(a.cfg.SessionsDir, name); err != nil {
			fmt.Printf("  ❌ %s\n", err)
		} else {
			fmt.Printf("  🗑️  Sesión '%s' borrada.\n", name)
			if a.sessionName == name {
				a.sessionName = ""
			}
		}
		return true

	case "/tools":
		fmt.Printf("\n%s🔧 Herramientas disponibles:%s\n\n", green, reset)
		for _, t := range a.registry.All() {
			fmt.Printf("  %s→%s %-15s %s\n", green, reset, t.Name(), t.Description())
		}
		fmt.Println()
		return true

	case "/limpiar":
		a.resetHistory()
		a.sessionName = ""
		fmt.Println("🧹 Historial y sesión activa limpiados.")
		return true

	case "/ayuda":
		autonomoStatus := "desactivado"
		if a.cfg.Autonomous {
			autonomoStatus = "✅ activado"
		}
		providerStatus := a.cfg.Provider
		if providerStatus == "" {
			providerStatus = "ollama"
		}
		fmt.Printf(`
📖 Comandos disponibles:
  /perfil                → muestra el perfil activo
  /perfil <nombre>       → cambia de perfil (pide confirmación)
  /sesion <nombre>       → carga o crea una sesión
  /sesion                → muestra la sesión activa
  /sesiones              → lista todas las sesiones guardadas
  /guardar               → guarda la sesión actual
  /borrar <nombre>       → borra una sesión
  /tools                 → lista las herramientas disponibles
  /limpiar               → limpia historial y sesión activa
  /ayuda                 → muestra esta ayuda
  salir / exit / quit    → cierra DeyaClaw

  Proveedor activo: %s
  Modo autónomo: %s
`, providerStatus, autonomoStatus)
		return true

	default:
		fmt.Printf("❓ Comando desconocido: %s — escribí /ayuda para ver los disponibles.\n", cmd)
		return true
	}
}

func (a *Agent) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	modeLabel := a.cfg.Mode
	if a.cfg.Autonomous {
		modeLabel += " ⚡autónomo"
	}

	providerLabel := a.cfg.Provider
	if providerLabel == "" {
		providerLabel = "ollama"
	}

	fmt.Printf("\n🦞 DeyaClaw — Modo: %s | Proveedor: %s\n", modeLabel, providerLabel)
	fmt.Println("Escribí tu consulta (o 'salir' para cerrar). /ayuda para comandos.")
	fmt.Println(strings.Repeat("─", 50))

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		if input == "salir" || input == "exit" || input == "quit" {
			if a.sessionName != "" {
				fmt.Printf("\n  ⚠️  Hay una sesión activa: %s\n", a.sessionName)
				fmt.Print("  ¿Guardás antes de salir? (s/n): ")
				if scanner.Scan() {
					if ans := strings.TrimSpace(strings.ToLower(scanner.Text())); ans == "s" || ans == "si" || ans == "sí" {
						a.saveSession()
					}
				}
			}
			fmt.Println("👋 Cerrando DeyaClaw.")
			break
		}

		if strings.HasPrefix(input, "/") {
			a.handleCommand(input, scanner)
			continue
		}

		if isSelfQuery(input) {
			fmt.Println("\n🤖 Soy DeyaClaw, un agente de ciberseguridad local. No tengo acceso a mi propio código fuente.")
			fmt.Println(strings.Repeat("─", 50))
			continue
		}

		a.history = append(a.history, ollama.Message{
			Role:    "user",
			Content: input,
		})

		response, err := a.chat(scanner)
		if err != nil {
			fmt.Printf("❌ Error: %s\n", err)
			a.history = a.history[:len(a.history)-1]
			continue
		}

		a.history = append(a.history, ollama.Message{
			Role:    "assistant",
			Content: response,
		})

		cyan  := "\033[36m"
		reset := "\033[0m"
		bold  := "\033[1m"

		fmt.Printf("\n🤖 ")
		lines := strings.Split(response, "\n")
		for i, line := range lines {
			if strings.Contains(line, "CVE-") {
				line = strings.ReplaceAll(line, "CVE-", bold+cyan+"CVE-"+reset)
			}
			if i == 0 {
				fmt.Print(line + "\n")
			} else {
				fmt.Println(line)
			}
		}
		fmt.Println(strings.Repeat("─", 50))

	}
}

func (a *Agent) SingleShot(query string) {
	green  := "\033[32m"
	bright := "\033[92m"
	reset  := "\033[0m"

	if isSelfQuery(query) {
		fmt.Println("🤖 Soy DeyaClaw, un agente de ciberseguridad local. No tengo acceso a mi propio código fuente.")
		return
	}

	a.history = append(a.history, ollama.Message{
		Role:    "user",
		Content: query,
	})

	scanner := bufio.NewScanner(os.Stdin)
	response, err := a.chat(scanner)
	if err != nil {
		fmt.Printf("❌ Error: %s\n", err)
		return
	}

	fmt.Printf("\n%s╔══════════════════════════════════════════════════╗%s\n", green, reset)
	fmt.Printf("%s║%s 🤖 DeyaClaw [%s%s%s]\n", green, reset, bright, a.cfg.Mode, reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════╝%s\n", green, reset)
	fmt.Println()
	fmt.Println(response)
	fmt.Println()
}
