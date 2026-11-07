package scaffold

import (
	"fmt"
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/services"
	"romanctl/internal/ui"
	"slices"
	"strings"
)

type catalogOutput struct {
	Stacks     []string                `json:"stacks"`
	Arch       []string                `json:"arch"`
	Services   []string                `json:"services"`
	Templates  []*Template             `json:"templates"`
	Simple     []map[string]any        `json:"simpleTypes"`
	Generators map[string][]*Generator `json:"generators"`
}

func Catalog() error {
	templates := LoadTemplates()
	out := catalogOutput{Arch: knownArch, Templates: templates}
	for _, t := range templates {
		if !slices.Contains(out.Stacks, t.Stack) {
			out.Stacks = append(out.Stacks, t.Stack)
		}
	}
	for _, s := range services.All {
		out.Services = append(out.Services, s.Key)
	}
	for _, t := range projects.ProjectTypes {
		out.Simple = append(out.Simple, map[string]any{"type": t.Key, "description": t.Label})
	}
	out.Generators = map[string][]*Generator{}
	for _, target := range AllTargets() {
		out.Generators[target] = LoadGenerators(target)
	}
	if ui.JSONOutput {
		return ui.PrintJSON(out)
	}
	fmt.Println(ui.STitle.Render(i18n.T("Templates de arquitetura")))
	for _, t := range templates {
		i18n.Printf("  %s  %s\n", ui.SBoldTxt.Render(t.Key()), ui.SGray.Render("("+t.Source+")"))
		fmt.Printf("    %s\n", i18n.T(t.Description))
		i18n.Printf("    romanctl new <nome> --stack %s --arch %s", t.Stack, strings.Join(t.Arch, ","))
		if len(t.Services) > 0 {
			i18n.Printf(" [--with %s]", strings.Join(t.Services, ","))
		}
		fmt.Println()
	}
	fmt.Println("\n" + ui.STitle.Render(i18n.T("Projetos simples")))
	fmt.Println(i18n.T("  romanctl new <tipo> <nome>   tipos: ") + projects.TypeKeys())
	fmt.Println("\n" + ui.STitle.Render(i18n.T("Geradores")) + ui.SGray.Render("  romanctl add <tipo> <nome>"))
	for _, target := range AllTargets() {
		fmt.Printf("  %-24s %s\n", target, kinds(LoadGenerators(target)))
	}
	i18n.Printf("\nTemplates próprios: %s/<stack>/<id>/template.yaml\n", userTemplatesDir())
	return nil
}
