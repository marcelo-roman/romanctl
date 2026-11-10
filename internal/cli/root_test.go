package cli

import (
	"reflect"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
	"strings"
	"testing"
)

func TestSplitFlags(t *testing.T) {
	flagError, flagVals = nil, map[string]string{}
	args, flags := splitFlags([]string{"new", "loja", "--stack", "dotnet", "--arch=clean,ddd", "--json", "-y"})
	if flagError != nil {
		t.Fatal(flagError)
	}
	if !reflect.DeepEqual(args, []string{"new", "loja"}) {
		t.Fatalf("args: %v", args)
	}
	if !flags["json"] || !flags["y"] || flagVals["stack"] != "dotnet" || flagVals["arch"] != "clean,ddd" {
		t.Fatalf("opções: %v %v", flags, flagVals)
	}
}

func TestSplitFlagsRecusaOpcaoInvalida(t *testing.T) {
	for _, in := range [][]string{{"--wiht", "postgres"}, {"--with postgres"}, {"--stack"}} {
		flagError, flagVals = nil, map[string]string{}
		splitFlags(in)
		if flagError == nil {
			t.Errorf("aceitou %v", in)
		}
	}
}

func TestPathDoUsuarioContinuaMandando(t *testing.T) {
	if got := pathWithBin("/a:/b", "/bin-romanctl"); got != "/a:/b:/bin-romanctl" {
		t.Fatal(got)
	}
	if got := pathWithBin("/a:/bin-romanctl:/b", "/bin-romanctl"); got != "/a:/bin-romanctl:/b" {
		t.Fatal(got)
	}
	if got := pathWithBin("", "/bin-romanctl"); got != "/bin-romanctl" {
		t.Fatal(got)
	}
}

func TestCategoriaUnicaNaoEntraNaBusca(t *testing.T) {
	one := []ui.Option{{Value: "a", Category: tools.GroupK8s, Search: "Kubernetes"}, {Value: "b", Category: tools.GroupK8s, Search: "Kubernetes"}}
	if got := toolCategories(one); got != nil {
		t.Fatalf("uma categoria não precisa de abas: %v", got)
	}
	if one[0].Search != "" || one[1].Search != "" {
		t.Fatalf("%+v", one)
	}
	two := []ui.Option{{Value: "a", Category: tools.GroupK8s, Search: "Kubernetes"}, {Value: "b", Category: tools.GroupDB, Search: "Banco"}}
	if got := toolCategories(two); len(got) != 2 || two[0].Search == "" {
		t.Fatalf("%v %+v", got, two)
	}
}

func TestPararODockerSemTerminalPedeYes(t *testing.T) {
	err := confirmWarning("docker", map[string]bool{})
	if err == nil || !ui.IsUsage(err) || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("sem terminal e sem --yes, parar o Docker deveria ser recusado: %v", err)
	}
	if err := confirmWarning("docker", map[string]bool{"yes": true}); err != nil {
		t.Fatal(err)
	}
	if err := confirmWarning("postgres", map[string]bool{}); err != nil {
		t.Fatal("serviço sem aviso não pede confirmação")
	}
}
