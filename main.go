// romanctl — utilitário de desenvolvimento para o tablet (Ubuntu em proot/Termux).
package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

const version = "0.1.0"

func usage() {
	fmt.Println(sTitle.Render("romanctl") + " " + sGray.Render("v"+version) + " — utilitário de desenvolvimento do tablet")
	fmt.Print(`
Uso: romanctl [comando] [argumentos]
     romanctl              abre o menu interativo

Serviços (code, postgres, mongo, redis, all):
  status                   mostra o estado de todos os serviços
  start   <serviço|all>    inicia
  stop    <serviço|all>    para
  restart <serviço|all>    reinicia
  restore [status|save|clear]  sobe de novo o que estava ligado (depois de uma queda)
  logs    <serviço>        acompanha o log (Ctrl+C para sair)
  install <serviço>        instala (precisa de internet)

Projetos (em ` + projDir + `):
  new [tipo] [nome]        cria projeto (sem argumentos: interativo)
  list                     lista projetos
  run <nome> [--open]      roda o projeto (npm run dev / dotnet run / go run)
                           --open abre o navegador quando o app subir
  code [projeto|pasta]     abre no code-server (sem alvo: projeto da pasta atual ou lista)
  open [alvo] [--app]      abre no navegador: code | <projeto> | <url>
                           <projeto> abre no code-server; --app abre o app rodando
  rm  <nome> [-y]          remove projeto (pede confirmação sem -y)
  types                    lista os tipos de projeto

Outros:
  doctor                   versões das ferramentas instaladas
  help                     esta ajuda

Offline (para ficar sem internet, como num voo):
  offline                  mostra o estado da preparação
  offline prep [--force]   baixa templates, caches e extensões (precisa de internet)
  offline test             cria e compila cada tipo com a rede bloqueada

Opções de new: --sem-git (não cria repositório git)
               --offline (usa só templates e cache; sem internet já é automático)
`)
}

func main() {
	args, flags := splitFlags(os.Args[1:])
	var err error
	if len(args) == 0 {
		err = menu()
	} else {
		err = dispatch(args, flags)
	}
	if isAbort(err) {
		infof("Cancelado")
		return
	}
	if err != nil {
		failf("%v", err)
		os.Exit(1)
	}
}

func splitFlags(in []string) (args []string, flags map[string]bool) {
	flags = map[string]bool{}
	for _, a := range in {
		if strings.HasPrefix(a, "-") {
			flags[strings.TrimLeft(a, "-")] = true
		} else {
			args = append(args, a)
		}
	}
	return
}

func arg(args []string, i int) string {
	if i < len(args) {
		return args[i]
	}
	return ""
}

func dispatch(args []string, flags map[string]bool) error {
	if flags["h"] || flags["help"] {
		usage()
		return nil
	}
	switch args[0] {
	case "help", "ajuda":
		usage()
	case "version", "versao":
		fmt.Println("romanctl", version)
	case "status", "st":
		printStatus()
	case "start", "iniciar", "up":
		return eachService(arg(args, 1), (*Service).Start)
	case "stop", "parar", "down":
		return eachService(arg(args, 1), (*Service).Stop)
	case "restart", "reiniciar":
		return eachService(arg(args, 1), func(s *Service) error {
			if err := s.Stop(); err != nil {
				return err
			}
			return s.Start()
		})
	case "install", "instalar":
		return eachService(arg(args, 1), (*Service).Install)
	case "restore", "restaurar":
		return restoreCommand(arg(args, 1))
	case "logs", "log":
		s := findService(arg(args, 1))
		if s == nil {
			return fmt.Errorf("informe o serviço: code, postgres, mongo ou redis")
		}
		return tailLog(s)
	case "new", "novo", "create":
		if arg(args, 1) == "" || arg(args, 2) == "" {
			return menuNewProject(arg(args, 1), arg(args, 2))
		}
		return newProject(args[1], args[2], !flags["sem-git"], flags["offline"])
	case "list", "ls", "projetos":
		printProjects()
	case "types", "tipos":
		for _, t := range projectTypes {
			fmt.Printf("  %-16s %s\n", t.Key, t.Label)
		}
	case "rm", "remove", "remover", "delete":
		if arg(args, 1) == "" {
			return menuRemoveProject()
		}
		if !flags["y"] && !flags["yes"] {
			ok, err := confirm(fmt.Sprintf("Remover %s/%s? Isso apaga TODOS os arquivos.", projDir, args[1]))
			if err != nil || !ok {
				infof("Cancelado")
				return nil
			}
		}
		return removeProject(args[1])
	case "run", "dev", "rodar":
		if arg(args, 1) == "" {
			return menuRunProject()
		}
		return runProject(args[1], flags["open"])
	case "code", "editar", "edit":
		return openInCode(arg(args, 1))
	case "open", "abrir":
		if arg(args, 1) == "" {
			return menuOpen()
		}
		return openTarget(args[1], flags["app"])
	case "offline":
		return offlineCommand(arg(args, 1), flags)
	case "doctor", "info":
		doctor()
	default:
		usage()
		return fmt.Errorf("comando desconhecido: %s", args[0])
	}
	return nil
}

func eachService(name string, f func(*Service) error) error {
	if name == "" {
		return fmt.Errorf("informe o serviço: code, postgres, mongo, redis ou all")
	}
	if name == "all" || name == "tudo" {
		var failed bool
		for _, s := range services {
			if !s.Installed() {
				continue
			}
			if err := f(s); err != nil {
				failf("%v", err)
				failed = true
			}
		}
		if failed {
			return fmt.Errorf("algum serviço falhou")
		}
		return nil
	}
	s := findService(name)
	if s == nil {
		return fmt.Errorf("serviço desconhecido %q (use code, postgres, mongo, redis ou all)", name)
	}
	return f(s)
}

func printStatus() {
	fmt.Println(sTitle.Render("Serviços"))
	for _, s := range services {
		var state string
		switch {
		case !s.Installed():
			state = sGray.Render("○ não instalado") + sGray.Render("  (romanctl install "+s.Key+")")
		case s.Running():
			state = sGreen.Render("● rodando") + "  " + s.URL()
		case wantsRunning(s.Key):
			state = sRed.Render("○ caiu") + sGray.Render("  (romanctl restore)")
		default:
			state = sRed.Render("○ parado")
		}
		fmt.Printf("  %-12s %s\n", s.Name, state)
	}
}

func tailLog(s *Service) error {
	if !exists(s.LogFile) {
		return fmt.Errorf("ainda não há log em %s", s.LogFile)
	}
	infof("%s  (Ctrl+C para sair)", s.LogFile)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	run("", nil, "tail", "-n", "50", "-f", s.LogFile)
	return nil
}

func doctor() {
	fmt.Println(sTitle.Render("Ferramentas"))
	tools := [][]string{
		{"node", "node", "--version"},
		{"npm", "npm", "--version"},
		{"go", "go", "version"},
		{".NET", "dotnet", "--version"},
		{"git", "git", "--version"},
		{"code-server", "code-server", "--version"},
		{"PostgreSQL", "/usr/lib/postgresql/16/bin/postgres", "--version"},
		{"MongoDB", mongoHome + "/bin/mongod", "--version"},
		{"mongosh", "mongosh", "--version"},
		{"Redis", "redis-server", "--version"},
	}
	for _, t := range tools {
		if b := pgBin(); t[0] == "PostgreSQL" && b != "" {
			t[1] = b + "/postgres"
		}
		v := sGray.Render("não instalado")
		if _, err := exec.LookPath(t[1]); err == nil {
			if out, err := output(t[1], t[2:]...); err == nil {
				v = strings.SplitN(out, "\n", 2)[0]
			}
		}
		fmt.Printf("  %-12s %s\n", t[0], v)
	}
	if out, err := output("sh", "-c", "free -h | awk '/Mem:/{print $7\" livres de \"$2}'"); err == nil {
		fmt.Printf("\n  %-12s %s\n", "Memória", out)
	}
	if out, err := output("sh", "-c", "df -h "+home+" | awk 'NR==2{print $4\" livres de \"$2}'"); err == nil {
		fmt.Printf("  %-12s %s\n", "Disco", out)
	}
	fmt.Printf("  %-12s %s\n  %-12s %s\n", "Projetos", projDir, "Dados", dataDir)
}
