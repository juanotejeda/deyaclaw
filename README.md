# 🦞 DeyaClaw

**Agente autónomo de ciberseguridad local**, powered by [Ollama](https://ollama.com) / [OpenRouter](https://openrouter.ai).

DeyaClaw es un agente de ciberseguridad que opera desde tu terminal, capaz de planificar y ejecutar
tareas ofensivas y defensivas de forma autónoma, encadenando herramientas de seguridad reales
(nmap, Metasploit, búsqueda de CVEs, etc.) guiadas por un LLM local o remoto.

---

## ✅ Estado actual (v0.1.1)

### Lo que ya funciona

- **Perfiles de agente**: RedTeam, BlueTeam, Docente, General.
- **Modo autónomo** (`-autonomo`): el agente encadena herramientas sin pedir confirmación para
  acciones seguras; solo pide confirmación explícita para `shell` y `metasploit`.
- **Pipeline RedTeam completo**:
  - `checkenv` → verifica herramientas disponibles.
  - `portscan` → escaneo de puertos con nmap (top100, all, custom).
  - `cvesearch` → búsqueda de CVEs por servicio/versión.
  - `exploitsearch` → búsqueda de exploits públicos (Exploit-DB, etc.).
  - `metasploit` → ejecución de módulos de Metasploit con opciones configurables.
  - `shell` → ejecución de comandos de sistema (con confirmación siempre).
  - `whois`, `websearch`, `whatweb` → reconocimiento pasivo y web.
- **Reporte de pentest**: el agente genera un reporte estructurado con tabla de hallazgos,
  severidad, CVE y recomendaciones al pedírselo (`"dame un reporte corto del pentest"`).
- **Memoria de sesión**: recuerda targets, puertos, versiones y CVEs encontrados en la sesión
  y los reutiliza como contexto.
- **Proveedor configurable**: Ollama (local) u OpenRouter (remoto).
- **Inferencia de parámetros**: `"mi PC"` → `127.0.0.1`, `"la red"` → `192.168.1.0/24`, etc.

### Tools disponibles

| Tool            | Descripción                                               |
|-----------------|-----------------------------------------------------------|
| `checkenv`      | Verifica el toolkit instalado (nmap, msfconsole, etc.)   |
| `portscan`      | Escaneo de puertos y detección de versiones con nmap     |
| `cvesearch`     | Búsqueda de CVEs por producto/versión                    |
| `exploitsearch` | Búsqueda de exploits públicos (Exploit-DB, etc.)         |
| `metasploit`    | Ejecución de módulos de Metasploit                       |
| `shell`         | Ejecución de comandos de sistema (requiere confirmación) |
| `whois`         | Consulta WHOIS de IPs y dominios                         |
| `websearch`     | Búsqueda web de información OSINT                        |
| `whatweb`       | Fingerprinting de servicios web                          |
| `sysinfo`       | Información del sistema local                            |

---

## 🚀 Instalación

### Requisitos

- Go 1.21+
- [Ollama](https://ollama.com) (para modelos locales) **o** API key de [OpenRouter](https://openrouter.ai)
- `nmap`
- `metasploit-framework` (`msfconsole`)
- `theHarvester` _(opcional)_
- `gobuster` _(opcional)_

### Build

```bash
git clone https://github.com/juanotejeda/deyaclaw.git
cd deyaclaw
go build -o deyaclaw .
Uso
bash
# Modo interactivo con perfil RedTeam y autonomía activada
./deyaclaw -perfil redteam -autonomo

# Modo BlueTeam sin autonomía
./deyaclaw -perfil blueteam

# Con proveedor OpenRouter
./deyaclaw -perfil redteam -autonomo -provider openrouter
Comandos internos
Comando	Acción
/ayuda	Muestra ayuda
/modo redteam	Cambia el perfil del agente
salir	Cierra el agente
🗺️ Roadmap
v0.2 — Explotación real en laboratorio
 Lab de pruebas documentado: guía para levantar Metasploitable2/3 o contenedor vulnerable
(vsftpd 2.3.4, Tomcat, etc.) para pruebas de explotación end-to-end.

 Flujo CVE → módulo Metasploit automático: cuando el agente encuentre un CVE con módulo
conocido, propondrá directamente el módulo a usar sin intervención del usuario.

 parseMsfOutput mejorado: interpretación correcta de banners, sesiones abiertas
(shell/meterpreter) y errores de Metasploit, sin perder información en el filtrado.

 Nivel de autonomía "full auto": en modo -autonomo, el agente sigue su propio plan
recomendado sin necesidad de que el usuario elija // entre pasos.

v0.3 — Reporte y persistencia
 Exportación de reportes: generar reportes en Markdown o PDF al finalizar una sesión
(/reporte pdf).

 Persistencia de sesión: guardar historial de hallazgos entre sesiones (SQLite o JSON).

 Modo scan continuo: monitoreo periódico de un target y alerta ante cambios
(nuevo puerto abierto, nueva versión, nuevo CVE publicado).

v0.4 — Nuevas tools y perfiles
 Tool: nikto → análisis de vulnerabilidades web.

 Tool: sqlmap → detección y explotación de inyecciones SQL.

 Tool: hydra → ataques de fuerza bruta a servicios (SSH, FTP, HTTP).

 Perfil forense → análisis de logs, detección de IOCs, análisis de artefactos.

 Perfil osint → reconocimiento pasivo avanzado (theHarvester, Shodan, Censys).

v0.5 — UX y autonomía avanzada
 Flag -objetivo: definir el target directamente desde la línea de comandos.

 Flag -plan: ejecutar un plan de fases completo sin intervención humana.

 Soporte multi-target: gestionar múltiples IPs/rangos en una sola sesión.

 Modo servidor (API REST): exponer el agente como servicio para integraciones externas.

🧠 Estado de desarrollo (contexto interno)
Esta sección documenta el estado técnico real del proyecto para mantener continuidad
entre sesiones de desarrollo.

Arquitectura actual
text
deyaclaw/
├── main.go               # Entry point, flags (-perfil, -autonomo, -provider)
├── agent/agent.go        # Loop principal (chat, confirmAction, isSafeTool)
├── config/config.go      # Configuración (Autonomous, Mode, Provider)
├── executor/executor.go  # Parsea JSON del LLM → ejecuta la Tool correspondiente
├── session/session.go    # Memoria de sesión (hallazgos, targets, contexto)
├── ollama/client.go      # Cliente HTTP para Ollama (local)
├── openrouter/client.go  # Cliente HTTP para OpenRouter (remoto)
└── tools/
    ├── tool.go           # Interface Tool {Name, Description, Execute}
    ├── checkenv.go       # Verifica toolkit instalado
    ├── portscan.go       # nmap wrapper
    ├── nmap.go           # nmap directo (alternativo)
    ├── cvesearch.go      # Búsqueda de CVEs
    ├── exploitsearch.go  # Búsqueda de exploits
    ├── metasploit.go     # msfconsole wrapper con parseMsfOutput
    ├── shell.go          # Ejecución de comandos del sistema
    ├── whois.go          # Consultas WHOIS
    ├── websearch.go      # Búsqueda web OSINT
    ├── sysinfo.go        # Info del sistema local
    └── sanitize.go       # Sanitización de inputs
Lógica de autonomía (agent.go)
isSafeTool(action string) bool: devuelve false solo para shell y metasploit.
Todo lo demás se ejecuta sin confirmación en modo autónomo.

confirmAction(preview string) bool: muestra ⚡ DeyaClaw quiere ejecutar: y acepta
s/si/sí como confirmación.

En modo -autonomo: las tools seguras se ejecutan directo; shell y metasploit
siguen pidiendo confirmación explícita al usuario.

Problema conocido activo: parseMsfOutput
parseMsfOutput en tools/metasploit.go filtra demasiado agresivamente el output de
msfconsole. Las líneas con [*] (informativas de Metasploit) son descartadas porque
el regex de important no las captura.

Próximo fix: ampliar el regex para incluir líneas [*] que contengan datos relevantes
(versión, fingerprint, IP, Scanned, completed).

Integración Metasploit
El modelo a veces manda options como map[RHOSTS:127.0.0.1 RPORT:22] en lugar
de string "RHOSTS=127.0.0.1 RPORT=22". Se instruyó al modelo en el prompt para
usar siempre formato string.

Execute en metasploit.go parsea el string con strings.Fields +
strings.SplitN(opt, "=", 2).

El módulo auxiliary/scanner/ssh/ssh_version funciona correctamente invocado directo
desde msfconsole, pero el output con [*] no es capturado por el filtro actual.

⚠️ Disclaimer
DeyaClaw está diseñado para uso en entornos de laboratorio controlados y sistemas propios.
El uso de este software contra sistemas sin autorización explícita es ilegal y va en contra
de la ética del hacking responsable. Los autores no se responsabilizan por el uso indebido
de esta herramienta.

📄 Licencia
MIT

