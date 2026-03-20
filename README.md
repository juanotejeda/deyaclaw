	# 🦞 DeyaClaw

**Agente autónomo de ciberseguridad local**, powered by [Ollama](https://ollama.com) / [OpenRouter](https://openrouter.ai).

DeyaClaw es un agente de ciberseguridad que opera desde tu terminal y coordina herramientas reales
(nmap, Metasploit, búsqueda de CVEs, etc.) guiado por un LLM local o remoto. No es “magia negra”:
su objetivo es ayudarte a **pensar y ejecutar mejor** en laboratorios y entornos controlados,
no “hackear solo”.

---

## ✅ Estado actual (v0.1.1)

### Qué hace hoy, de forma razonablemente estable

- **Perfiles de agente**:  
  - `redteam` → ofensivo, orientado a encontrar y explotar vulnerabilidades.
  - `blueteam` → defensivo, orientado a exposición, riesgo y mitigaciones.
  - `docente` → explica cada paso y resultado.
  - `general` → modo genérico.

- **Modo autónomo** (`-autonomo`):  
  - Encadena herramientas “seguras” sin pedir confirmación.  
  - Siempre pide confirmación explícita para acciones sensibles (`shell` y `metasploit`).

- **Pipeline RedTeam básico pero funcional**:
  - `checkenv` → verifica qué herramientas están instaladas (nmap, msfconsole, etc.).
  - `portscan` → escaneo de puertos/servicios con nmap (top100, all, custom).
  - `cvesearch` → búsqueda de CVEs por servicio/versión.
  - `exploitsearch` → búsqueda de exploits públicos (Exploit-DB, etc.).
  - `metasploit` → ejecución de módulos de Metasploit con parámetros configurables.
  - `shell` → ejecución de comandos del sistema (siempre con confirmación).
  - `whois`, `websearch` → reconocimiento pasivo / OSINT.
  - `sysinfo` → contexto del sistema local al inicio de la sesión.

- **Reporte de pentest** (MVP):  
  El agente puede generar un **reporte corto estructurado** cuando se le pide
  (por ejemplo: `"dame un reporte corto del pentest"`), con:
  - Resumen ejecutivo.
  - Tabla de hallazgos (título, severidad, servicio/PUERTO, CVE).
  - Detalle por hallazgo y recomendaciones.

- **Memoria de sesión**:
  - Recuerda targets, puertos, servicios, versiones y CVEs vistos en la sesión.
  - Reutiliza ese contexto para siguientes pasos sin que tengas que repetirlo.

- **Proveedor de modelo configurable**:
  - [Ollama](https://ollama.com) para modelos locales.
  - [OpenRouter](https://openrouter.ai) para modelos remotos.

- **Inferencia de parámetros** (callejera pero útil):
  - `"mi PC"` / `"localhost"` / `"mi máquina"` → `127.0.0.1`.
  - `"la red"` / `"red local"` → `192.168.1.0/24`.
  - Sin puertos → `top100`.
  - `"todos los puertos"` → scan completo.
  - CVE sin prefijo → agrega automáticamente `"CVE-"` si tiene formato año-id.

---

## 🔧 Tools disponibles

Hoy el agente puede llamar a estas herramientas internas:

| Tool            | Descripción                                                      |
|-----------------|------------------------------------------------------------------|
| `checkenv`      | Verifica el toolkit instalado (nmap, msfconsole, etc.)          |
| `portscan`      | Escaneo de puertos / versiones con nmap                         |
| `cvesearch`     | Búsqueda de CVEs por producto/versión                            |
| `exploitsearch` | Búsqueda de exploits públicos (Exploit-DB, etc.)                |
| `metasploit`    | Wrapper para ejecutar módulos de Metasploit (`msfconsole`)      |
| `shell`         | Ejecución de comandos de sistema (requiere confirmación)        |
| `whois`         | Consultas WHOIS de IPs y dominios                               |
| `websearch`     | Búsqueda web/OSINT (resúmenes de resultados)                    |
| `sysinfo`       | Información del sistema local (contexto inicial)               |

La lista crecerá, pero el criterio es mantener **pocas tools bien integradas**, en lugar de una
lista enorme que no se usa.

---

## 🔗 Integración con Metasploit (beta honesta)

DeyaClaw no “sabe explotar todo”, pero ya puede ayudarte de forma útil con Metasploit:

### Qué hace bien

- Buscar exploits públicos relacionados a un CVE usando Exploit-DB.
- Detectar si existe un módulo de Metasploit para ese CVE y mostrar el nombre (`exploit/...`).
- Ejecutar módulos de Metasploit (auxiliary y exploit) desde lenguaje natural,
  pidiendo confirmación previa.
- Resumir el output de Metasploit en lenguaje claro (sin pegar la consola completa).

> Nota: hoy el foco de la integración con Metasploit es **ayudarte a llegar rápido al módulo correcto y ejecutar comandos bien formados**, no “hackear solo”. Para exploits más complejos sigue siendo necesario ajustar manualmente algunos parámetros (shares, usuarios, payloads, etc.).

### Limitaciones (importante leer)

- Exploits complejos (por ejemplo, `exploit/linux/samba/is_known_pipename` para **CVE‑2017‑7494**)
  siguen necesitando que el usuario ajuste manualmente parámetros específicos como:
  - `SMB_SHARE_NAME` (share escribible).
  - Credenciales (`SMBUser`, `SMBPass`).
  - Payloads u opciones avanzadas.
- La lógica de “primero enumero, después exploto” (por ejemplo, enumerar shares SMB antes de lanzar
  un exploit) está descrita en el prompt y se está iterando; **no es 100% fiable ni general** aún.
- El enfoque actual es: **ayudar a llegar rápido al módulo correcto y ejecutar comandos bien
  formados**, no sustituir al operador humano.

### Requisitos para Metasploit

- Metasploit Framework instalado y accesible como `msfconsole`.
- Módulos disponibles en una ruta estándar (por ejemplo:
  `/opt/metasploit-framework/embedded/framework/modules`).

---

## 🚀 Instalación

### Requisitos

- Go 1.21+
- [Ollama](https://ollama.com) **o** API key de [OpenRouter](https://openrouter.ai)
- `nmap`
- `metasploit-framework` (`msfconsole`)
- `theHarvester` _(opcional, si querés ampliar reconocimiento)_  
- `gobuster` _(opcional)_

### Build

```bash
git clone https://github.com/juanotejeda/deyaclaw.git
cd deyaclaw
go build -o deyaclaw .
```
### Perfiles por defecto
Para usar los mismos perfiles (redteam, blueteam, docente, general) con los que se desarrolla DeyaClaw, copiá los JSON incluidos en el repositorio a tu home:

```bash
mkdir -p ~/.deyaclaw/profiles
cp profiles/*.json ~/.deyaclaw/profiles/
```
Si no copiás estos perfiles, igual podés usar el flag -perfil, pero algunos comandos internos como /perfil redteam mostrarán un mensaje indicando que el perfil JSON no existe aún.

### Uso

```bash
# Modo interactivo con perfil RedTeam y autonomía activada
./deyaclaw -perfil redteam -autonomo

# Modo BlueTeam sin autonomía
./deyaclaw -perfil blueteam

# Con proveedor OpenRouter
./deyaclaw -perfil redteam -autonomo -provider openrouter
```

### Comandos internos

| Comando            | Acción                                         |
|--------------------|------------------------------------------------|
| `/ayuda`           | Muestra ayuda y estado actual                  |
| `/perfil`          | Muestra el perfil activo                       |
| `/perfil <perfil>` | Cambia el perfil (`redteam`, `blueteam`, etc.) |
| `salir`            | Cierra el agente                               |

---

## 🗺️ Roadmap

### v0.2 — Explotación real en laboratorio

- [ ] **Lab de pruebas documentado**: guía para levantar Metasploitable2/3 o contenedor vulnerable
  (vsftpd 2.3.4, Tomcat, etc.) para pruebas de explotación end-to-end.
- [ ] **Flujo CVE → módulo Metasploit automático**: cuando el agente encuentre un CVE con módulo
  conocido, propondrá directamente el módulo a usar sin intervención del usuario.
- [ ] **`parseMsfOutput` mejorado**: 
	- Interpretación correcta de banners
	- Sesiones abiertas (shell/meterpreter) 
	- Errores de Metasploit, sin perder información en el filtrado.
- [ ] **Nivel de autonomía "full auto"**: en modo `-autonomo`, seguir un plan completo sin menús intermedios, siempre con límites claros en acciones destructivas.

### v0.3 — Reporte y persistencia

- [ ] **Exportación de reportes**: generar reportes en Markdown o PDF al finalizar una sesión (`/reporte pdf`).
- [ ] **Persistencia de sesión**: guardar historial de hallazgos entre sesiones (SQLite o JSON).
- [ ] **Modo scan continuo**: monitoreo periódico de un target y alerta ante cambios
  (nuevo puerto abierto, nueva versión, nuevo CVE publicado).

### v0.4 — Nuevas tools y perfiles

- [ ] **Tool: `nikto`** → análisis de vulnerabilidades web.
- [ ] **Tool: `sqlmap`** → detección y explotación de inyecciones SQL.
- [ ] **Tool: `hydra`** → ataques de fuerza bruta a servicios (SSH, FTP, HTTP).
- [ ] **Perfil `forense`** → análisis de logs, detección de IOCs, análisis de artefactos.
- [ ] **Perfil `osint`** → reconocimiento pasivo avanzado (theHarvester, Shodan, Censys).

### v0.5 — UX y autonomía avanzada

- [ ] **Flag `-objetivo`**: definir el target directamente desde la línea de comandos.
- [ ] **Flag `-plan`**: ejecutar un plan de fases completo sin intervención humana.
- [ ] **Soporte multi-target**: gestionar múltiples IPs/rangos en una sola sesión.
- [ ] **Modo servidor (API REST)**: exponer el agente como servicio para integraciones externas.

---

## 🧠 Estado de desarrollo (contexto interno)

> Esta sección documenta el estado técnico real del proyecto para mantener continuidad
> entre sesiones de desarrollo.

### Arquitectura actual

```
deyaclaw/
├── main.go               # Entry point, flags (-perfil, -autonomo, -provider)
├── agent/
│   └── agent.go          # Loop principal, chat, confirmAction, sesiones
├── config/
│   └── config.go         # Config (Autonomous, Mode, Provider, etc.)
├── executor/
│   └── executor.go       # Parsea JSON del LLM → ejecuta la Tool correspondiente
├── session/
│   └── session.go        # Manejo de sesiones e historial
├── ollama/
│   └── client.go         # Cliente HTTP para Ollama (local)
├── openrouter/
│   └── client.go         # Cliente HTTP para OpenRouter (remoto)
├── tools/
│   ├── tool.go           # Interface Tool {Name, Description, Execute}
│   ├── checkenv.go       # Verifica toolkit instalado
│   ├── portscan.go       # Wrapper nmap
│   ├── nmap.go           # nmap directo (alternativa)
│   ├── cvesearch.go      # Búsqueda de CVEs
│   ├── exploitsearch.go  # Búsqueda de exploits públicos
│   ├── metasploit.go     # Wrapper msfconsole + parseMsfOutput
│   ├── shell.go          # Comandos de sistema
│   ├── websearch.go      # Búsqueda web OSINT
│   ├── whois.go          # WHOIS
│   ├── sysinfo.go        # Info del sistema local
│   └── sanitize.go       # Sanitización de outputs
└── scripts/
    ├── setup-pentest.sh      # (en progreso) helper para montar toolkit
    └── update_msf_modules.sh # helper para refrescar cache de módulos MSF
```

### Autonomía y seguridad

- `isSafeTool` marca como “sensibles” solo `shell` y `metasploit`.
  En modo `-autonomo`:
  - Tools seguras se ejecutan directo.
  - Tools sensibles siempre piden confirmación.
- El prompt del agente limita: 
	- Qué herramientas puede usar.
	- Qué comandos de sistema están permitidos.
	- Cómo debe encadenar herramientas sin pedirte que elijas cada vez.

### Problema conocido activo: parseMsfOutput

- `parseMsfOutput` todavía filtra más de la cuenta en algunos casos.
- Encadenamientos avanzados (ej. enumerar shares SMB y después explotar) todavía dependen mucho del modelo y requieren supervisión humana.
---

## ⚠️ Disclaimer

DeyaClaw está diseñado para uso en **laboratorios de práctica** y **sistemas propios o
explícitamente autorizados**.
El uso contra sistemas sin autorización es **ilegal** y contrario a la ética del hacking
responsable. Los autores no se responsabilizan por el uso indebido de esta herramienta.

---

## 🙌 Agradecimientos

Un agradecimiento especial a la comunidad de **#RemoteExecution** ([foro.remoteexecution.org](https://foro.remoteexecution.org)) por el feedback, las ideas y el testing en entornos reales de laboratorio.

---
## 📄 Licencia

MIT — uso, copia, modificación y distribución permitidos, con la condición
de mantener el aviso de copyright y el aviso de licencia.

---

Autor: [Juan O. Tejeda](https://github.com/juanotejeda)  
Si encontrás bugs o tenés ideas, abrí un issue o PR en el repo.
