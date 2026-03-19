package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/juano/deyaclaw/agent"
	"github.com/juano/deyaclaw/config"
)

const version = "0.1.1"

var matrixChars = []rune("アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%&")

func checkNotRoot() {
	yellow := "\033[33m"
	red    := "\033[31m"
	reset  := "\033[0m"

	u, err := user.Current()
	if err != nil {
		return
	}

	if u.Uid == "0" {
		fmt.Printf("\n%s╔══════════════════════════════════════════════════╗%s\n", red, reset)
		fmt.Printf("%s║  ⚠️  ADVERTENCIA: Estás ejecutando como root      ║%s\n", red, reset)
		fmt.Printf("%s║  DeyaClaw puede ejecutar comandos en tu sistema.  ║%s\n", red, reset)
		fmt.Printf("%s║  Se recomienda correr como usuario sin privilegios ║%s\n", red, reset)
		fmt.Printf("%s╚══════════════════════════════════════════════════╝%s\n", red, reset)
		fmt.Printf("\n%s¿Querés continuar de todas formas? (s/n): %s", yellow, reset)

		var ans string
		fmt.Scanln(&ans)
		ans = strings.TrimSpace(strings.ToLower(ans))
		if ans != "s" && ans != "si" && ans != "sí" {
			fmt.Println("👋 Saliendo.")
			os.Exit(0)
		}
	}
}
func matrixRain() {
	green  := "\033[32m"
	bright := "\033[92m"
	reset  := "\033[0m"
	dim    := "\033[2m"

	cols := 60
	drops := make([]int, cols)
	for i := range drops {
		drops[i] = rand.Intn(10)
	}

	for row := 0; row < 8; row++ {
		line := make([]string, cols)
		for col := 0; col < cols; col++ {
			ch := string(matrixChars[rand.Intn(len(matrixChars))])
			if drops[col] == row {
				line[col] = bright + ch + reset
			} else if drops[col] > row {
				line[col] = dim + green + ch + reset
			} else {
				line[col] = " "
			}
		}
		fmt.Println(strings.Join(line, ""))
		time.Sleep(40 * time.Millisecond)
	}
}

func printBanner() {
	green  := "\033[32m"
	bright := "\033[92m"
	yellow := "\033[33m"
	reset  := "\033[0m"
	dim    := "\033[2m"

	matrixRain()

	fmt.Println()
	fmt.Println(bright + `  ██████╗ ███████╗██╗   ██╗ █████╗  ██████╗██╗      █████╗ ██╗    ██╗` + reset)
	fmt.Println(bright + `  ██╔══██╗██╔════╝╚██╗ ██╔╝██╔══██╗██╔════╝██║     ██╔══██╗██║    ██║` + reset)
	fmt.Println(bright + `  ██║  ██║█████╗   ╚████╔╝ ███████║██║     ██║     ███████║██║ █╗ ██║` + reset)
	fmt.Println(bright + `  ██║  ██║██╔══╝    ╚██╔╝  ██╔══██║██║     ██║     ██╔══██║██║███╗██║` + reset)
	fmt.Println(bright + `  ██████╔╝███████╗   ██║   ██║  ██║╚██████╗███████╗██║  ██║╚███╔███╔╝` + reset)
	fmt.Println(bright + `  ╚═════╝ ╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝` + reset)
	fmt.Println()
	fmt.Printf("  %s[ Agente de Ciberseguridad Local ]%s\n", green, reset)
	fmt.Printf("  %sv%s%s   %sOllama powered%s   %s🦞%s\n", yellow, version, reset, dim, reset, bright, reset)
	fmt.Println()
}

type profileMeta struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	file  string
}

func loadProfilesMeta(profilesDir string) []profileMeta {
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil
	}
	var profiles []profileMeta
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			path := filepath.Join(profilesDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var meta profileMeta
			if err := json.Unmarshal(data, &meta); err != nil {
				continue
			}
			meta.file = strings.TrimSuffix(e.Name(), ".json")
			profiles = append(profiles, meta)
		}
	}
	return profiles
}

func selectProfile(profilesDir string) string {
	green  := "\033[32m"
	bright := "\033[92m"
	yellow := "\033[33m"
	dim    := "\033[2m"
	reset  := "\033[0m"

	profiles := loadProfilesMeta(profilesDir)

	fmt.Println(green + "  ┌─────────────────────────────────┐" + reset)
	fmt.Println(green + "  │   " + bright + "Seleccioná un perfil" + reset + green + "         │" + reset)
	fmt.Println(green + "  ├─────────────────────────────────┤" + reset)

	for i, p := range profiles {
		line := fmt.Sprintf("  │  %s[%d]%s  %s  %-24s│",
			yellow, i+1, reset, p.Emoji, p.Name)
		fmt.Println(green + line + reset)
	}

	fmt.Printf(green+"  │  %s[0]%s  ⚪  %-24s│\n"+reset, yellow, reset, "General")
	fmt.Println(green + "  └─────────────────────────────────┘" + reset)
	fmt.Println()
	fmt.Printf(dim+"  Ingresá el número: "+reset)

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "0" || input == "" {
		return ""
	}

	for i, p := range profiles {
		if input == fmt.Sprintf("%d", i+1) {
			return p.file
		}
	}

	fmt.Println("  ❌ Opción inválida, arrancando en modo general.")
	return ""
}

func listProfiles(profilesDir string) {
	green  := "\033[32m"
	reset  := "\033[0m"

	profiles := loadProfilesMeta(profilesDir)
	fmt.Printf("\n%s📋 Perfiles disponibles:%s\n\n", green, reset)
	if len(profiles) == 0 {
		fmt.Println("  No hay perfiles en", profilesDir)
		return
	}
	for _, p := range profiles {
		fmt.Printf("  %s→%s %s  %s\n", green, reset, p.Emoji, p.Name)
	}
	fmt.Println()
}

func main() {
	checkNotRoot()
	rand.Seed(time.Now().UnixNano())

	versionFlag  := flag.Bool("version", false, "muestra la versión")
	perfilFlag   := flag.String("perfil", "", "perfil a cargar: redteam, blueteam, profesor")
	perfilesFlag := flag.Bool("perfiles", false, "lista los perfiles disponibles")
	qFlag        := flag.String("q", "", "consulta directa sin modo interactivo (single-shot)")
	sesionFlag := flag.String("sesion", "", "nombre de sesión a cargar o crear")
	autonomoFlag := flag.Bool("autonomo", false, "modo autónomo sin confirmación de acciones")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("🦞 DeyaClaw v%s\n", version)
		os.Exit(0)
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".deyaclaw", "config.json")

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ Error cargando config: %s\n", err)
		os.Exit(1)
	}

	if *perfilesFlag {
		listProfiles(cfg.ProfilesDir)
		os.Exit(0)
	}

	printBanner()

	selectedPerfil := *perfilFlag
	if selectedPerfil == "" && *qFlag == "" {
		// Solo mostramos el selector en modo interactivo
		selectedPerfil = selectProfile(cfg.ProfilesDir)
	}

	if selectedPerfil != "" {
		if err := cfg.LoadProfile(selectedPerfil); err != nil {
			fmt.Printf("❌ %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n  🔰 Perfil cargado: %s\n", cfg.Mode)
	}

	if *autonomoFlag {
	    cfg.Autonomous = true
	}

	a := agent.New(cfg)

	if *qFlag != "" {
		a.SingleShot(*qFlag)
	} else {
		a.Run()
	}
	if *sesionFlag != "" {
    	a.LoadSession(*sesionFlag)
	}
}
