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
	args, flags := splitFlags([]string{"new", "shop", "--stack", "dotnet", "--arch=clean,ddd", "--json", "-y"})
	if flagError != nil {
		t.Fatal(flagError)
	}
	if !reflect.DeepEqual(args, []string{"new", "shop"}) {
		t.Fatalf("args: %v", args)
	}
	if !flags["json"] || !flags["y"] || flagVals["stack"] != "dotnet" || flagVals["arch"] != "clean,ddd" {
		t.Fatalf("options: %v %v", flags, flagVals)
	}
}

func TestSplitFlagsRejectsInvalidOption(t *testing.T) {
	for _, in := range [][]string{{"--wiht", "postgres"}, {"--with postgres"}, {"--stack"}} {
		flagError, flagVals = nil, map[string]string{}
		splitFlags(in)
		if flagError == nil {
			t.Errorf("accepted %v", in)
		}
	}
}

func TestUserPathStillWins(t *testing.T) {
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

func TestSingleCategoryStaysOutOfSearch(t *testing.T) {
	one := []ui.Option{{Value: "a", Category: tools.GroupK8s, Search: "Kubernetes"}, {Value: "b", Category: tools.GroupK8s, Search: "Kubernetes"}}
	if got := toolCategories(one); got != nil {
		t.Fatalf("a single category needs no tabs: %v", got)
	}
	if one[0].Search != "" || one[1].Search != "" {
		t.Fatalf("%+v", one)
	}
	two := []ui.Option{{Value: "a", Category: tools.GroupK8s, Search: "Kubernetes"}, {Value: "b", Category: tools.GroupDB, Search: "Databases"}}
	if got := toolCategories(two); len(got) != 2 || two[0].Search == "" {
		t.Fatalf("%v %+v", got, two)
	}
}

func TestStoppingDockerWithoutTerminalNeedsYes(t *testing.T) {
	err := confirmWarning("docker", map[string]bool{})
	if err == nil || !ui.IsUsage(err) || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("without a terminal and without --yes, stopping Docker should be refused: %v", err)
	}
	if err := confirmWarning("docker", map[string]bool{"yes": true}); err != nil {
		t.Fatal(err)
	}
	if err := confirmWarning("postgres", map[string]bool{}); err != nil {
		t.Fatal("a service without a warning does not ask for confirmation")
	}
}
