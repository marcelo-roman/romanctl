package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/offline"
	"romanctl/internal/platform"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/services"
	"romanctl/internal/state"
	"romanctl/internal/tui"
	"romanctl/internal/ui"
	"slices"
	"strings"
)

// version comes from the release tag (-ldflags "-X romanctl/internal/cli.version=..."); the value
// below applies to builds from source.
var version = "0.5.2"

// pathWithBin appends romanctl's folder: the user's PATH still wins, and a program
// they chose is never replaced by one romanctl installed.
func pathWithBin(current, bin string) string {
	if slices.Contains(filepath.SplitList(current), bin) {
		return current
	}
	if current == "" {
		return bin
	}
	return current + string(os.PathListSeparator) + bin
}

func Main() {
	if len(os.Args) > 1 && os.Args[1] == projects.SuperviseArg {
		os.Exit(projects.SuperviseHere())
	}
	if p := platform.Current; p.OS != "linux" && p.OS != "darwin" {
		ui.Failf("romanctl only works on Linux and macOS (here: %s)", p.OS)
		os.Exit(1)
	}
	// What romanctl installs must be found even outside the user's PATH.
	os.Setenv("ROMANCTL_ORIGINAL_PATH", os.Getenv("PATH"))
	os.Setenv("PATH", pathWithBin(os.Getenv("PATH"), platform.Current.BinDir))
	args, flags := splitFlags(os.Args[1:])
	ui.JSONOutput = flags["json"]
	scaffold.Version = version
	var err error
	if flags["help"] || flags["h"] {
		// Comes first: "--help" never runs the command, not even with invalid arguments.
		printHelp(args)
		return
	}
	if flagError != nil {
		err = flagError
	} else if len(args) == 0 {
		// Without a terminal there is no list: show the help instead of failing.
		if !ui.Interactive() {
			usage()
		} else if err = requireLogin(); err == nil {
			err = commandMenu()
		}
	} else if needsLogin(args) {
		if err = requireLogin(); err == nil {
			err = dispatch(args, flags)
		}
	} else {
		err = dispatch(args, flags)
	}
	if ui.IsAbort(err) {
		ui.Infof("Cancelled")
		return
	}
	if errors.Is(err, ui.ErrReported) {
		os.Exit(1)
	}
	if err != nil {
		if ui.JSONOutput {
			ui.PrintJSON(map[string]any{"ok": false, "error": err.Error(), "usage": ui.IsUsage(err)})
		} else {
			ui.Failf("%v", err)
		}
		if ui.IsUsage(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

var valueFlags = map[string]bool{"stack": true, "arch": true, "with": true, "seconds": true, "interval": true, "lang": true, "project": true, "in": true, "description": true, "filter": true, "category": true, "version": true, "ref": true, "plural": true, "port": true,
	"issuer": true, "api": true, "idp": true, "display-name": true, "email": true, "first-name": true, "last-name": true, "search": true,
	"credential": true, "roles": true, "global": true, "realms": true, "name": true, "limit": true, "data": true,
	"client-id": true, "token-url": true}

var flagVals = map[string]string{}

var boolFlags = map[string]bool{"json": true, "y": true, "yes": true, "force": true, "app": true,
	"open": true, "all": true, "public": true, "no-ci": true, "list": true, "plain": true, "d": true, "detach": true, "check": true, "no-run": true, "keep-services": true, "no-git": true, "offline": true, "dry-run": true, "no-readme": true, "h": true, "help": true}

// flagAliases keeps older flag spellings working.
var flagAliases = map[string]string{"sem-git": "no-git"}

var flagError error

func splitFlags(in []string) (args []string, flags map[string]bool) {
	flags = map[string]bool{}
	for i := 0; i < len(in); i++ {
		a := in[i]
		if !strings.HasPrefix(a, "-") {
			args = append(args, a)
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(a, "-"), "=")
		if alias, ok := flagAliases[name]; ok {
			name = alias
		}
		switch {
		case !valueFlags[name] && !boolFlags[name]:
			flagError = ui.Usagef("unknown option: %s", a)
		case !valueFlags[name]:
			flags[name] = true
		case hasValue:
			flagVals[name] = value
		case i+1 < len(in):
			i++
			flagVals[name] = in[i]
		default:
			flagError = ui.Usagef("option --%s needs a value", name)
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
	switch args[0] {
	case "help", "ajuda":
		printHelp(args[1:])
	case "version", "versao":
		fmt.Println("romanctl", version)
	case "login", "entrar":
		return loginCommand()
	case "logout", "sair":
		return logoutCommand()
	case "whoami":
		return whoamiCommand()
	case "iam", "keycloak":
		return iamCommand(args[1:], flags)
	case "status", "st":
		if ui.JSONOutput {
			return ui.PrintJSON(services.StatusJSON(flags["all"]))
		}
		services.PrintStatus(flags["all"])
	case "catalog", "catalogo":
		return scaffold.Catalog()
	case "start", "iniciar":
		return serviceCommand("start", arg(args, 1), (*services.Service).Start)
	case "stop", "parar":
		if err := confirmWarning(arg(args, 1), flags); err != nil {
			return err
		}
		return serviceCommand("stop", arg(args, 1), (*services.Service).Stop)
	case "restart", "reiniciar":
		if err := confirmWarning(arg(args, 1), flags); err != nil {
			return err
		}
		return serviceCommand("restart", arg(args, 1), func(s *services.Service) error {
			if err := s.Stop(); err != nil {
				return err
			}
			return s.Start()
		})
	case "install", "instalar":
		return serviceCommand("install", arg(args, 1), (*services.Service).Install)
	case "restore", "restaurar":
		if arg(args, 1) == "" && ui.Interactive() {
			return menuRestore()
		}
		return services.RestoreCommand(arg(args, 1))
	case "logs", "log":
		if arg(args, 1) == "" && ui.Interactive() {
			return menuLogs()
		}
		s := services.FindService(arg(args, 1))
		if s == nil {
			// Also follows the log of a project brought up with up -d.
			if e, ok := state.ReadState().Ups[arg(args, 1)]; ok && e.Log != "" {
				return services.TailFile(e.Log)
			}
			return ui.Usagef("specify the service (see: romanctl status)")
		}
		return services.TailLog(s)
	case "new", "novo", "create":
		return newCommand(args[1:], flags)
	case "list", "ls", "projetos":
		return projectList(flags["plain"])
	case "types", "tipos":
		for _, t := range projects.ProjectTypes {
			if filter := flagVals["filter"]; filter == "" || ui.Matches(t.Key+" "+t.Label, filter) {
				i18n.Printf("  %-16s %s\n", t.Key, t.Label)
			}
		}
	case "rm", "remove", "remover", "delete":
		if arg(args, 1) == "" {
			return menuRemoveProject()
		}
		if !flags["y"] && !flags["yes"] {
			ok, err := ui.Confirm(i18n.Sprintf("Remove %s/%s? This deletes ALL files.", config.ProjDir, args[1]))
			if err != nil || !ok {
				ui.Infof("Cancelled")
				return nil
			}
		}
		return projects.RemoveProject(args[1])
	case "run", "dev", "rodar":
		if arg(args, 1) == "" {
			return menuRunProject()
		}
		return projects.RunProject(args[1], flags["open"])
	case "code", "editar", "edit":
		return openInCode(arg(args, 1))
	case "open", "abrir":
		if arg(args, 1) == "" {
			return menuOpen()
		}
		return openTarget(args[1], flags["app"])
	case "offline":
		if arg(args, 1) == "" && ui.Interactive() {
			return menuOffline()
		}
		return offline.Command(arg(args, 1), flags)
	case "tools", "ferramentas":
		return toolsCommand(args[1:], flags)
	case "web-ui", "webui", "web":
		return webUICommand(flags)
	case "templates", "template":
		return templatesCommand(args[1:])
	case "claude", "skills":
		return claudeCommand(args[1:], flags)
	case "system", "sistema", "sys":
		return systemCommand(args[1:])
	case "up", "subir":
		return upCommand(args[1:], flags)
	case "down", "derrubar":
		return downCommand(args[1:], flags)
	case "ui", "painel", "tui":
		if !ui.Interactive() {
			return ui.Usagef("the panel needs a terminal")
		}
		return tui.Run(version)
	case "menu":
		return menu()
	case "self-update", "update", "atualizar":
		return selfUpdateCommand(flags)
	case "config":
		return configCommand(args[1:])
	case "add", "adicionar", "generate", "g":
		return addCommand(args[1:], flags)
	case "publish", "publicar":
		return publishCommand(args[1:], flags)
	case "platform", "plataforma":
		if ui.JSONOutput {
			return ui.PrintJSON(platform.Current)
		}
		fmt.Println(platform.Current.Describe())
		return nil
	case "doctor", "info":
		doctor()
	default:
		return ui.Usagef("unknown command: %s (see: romanctl help)", args[0])
	}
	return nil
}

func eachService(name string, f func(*services.Service) error) error {
	if name == "" {
		return ui.Usagef("specify the service (see: romanctl status) or all")
	}
	if name == "all" || name == "tudo" {
		var failed bool
		for _, s := range services.All {
			// The engine stays out: stopping Docker would bring down containers that are not romanctl's.
			if s.Unsupported != "" || s.Engine || !s.Installed() {
				continue
			}
			if err := f(s); err != nil {
				ui.Failf("%v", err)
				failed = true
			}
		}
		if failed {
			return i18n.Errorf("some service failed")
		}
		return nil
	}
	s := services.FindService(name)
	if s == nil {
		return ui.Usagef("unknown service %q (see: romanctl status)", name)
	}
	return f(s)
}

// confirmWarning asks for confirmation before stopping a service that takes other things down with it,
// such as Docker. Without a terminal, it only proceeds with --yes.
func confirmWarning(name string, flags map[string]bool) error {
	s := services.FindService(name)
	if s == nil || s.Warning == "" || flags["y"] || flags["yes"] {
		return nil
	}
	if !ui.Interactive() {
		return ui.Usagef("%s: %s; confirm with --yes", s.Name, s.Warning)
	}
	ok, err := ui.Confirm(i18n.Sprintf("%s: %s. Continue?", s.Name, s.Warning))
	if err != nil {
		return err
	}
	if !ok {
		return ui.ErrCanceled
	}
	return nil
}
