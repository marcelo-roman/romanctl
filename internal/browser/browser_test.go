package browser

import (
	"strings"
	"testing"

	"romanctl/internal/platform"
)

func TestOpener(t *testing.T) {
	only := func(bins ...string) func(string) bool {
		return func(b string) bool { return strings.Contains(strings.Join(bins, " "), b) }
	}
	cases := []struct {
		p    platform.Info
		have func(string) bool
		want string
	}{
		{platform.Info{OS: "darwin"}, only(), "open"},
		{platform.Info{OS: "linux"}, only(termuxPrefix + "/bin/termux-open-url"), termuxPrefix + "/bin/termux-open-url"},
		{platform.Info{OS: "linux", WSL: true}, only("wslview", "xdg-open"), "wslview"},
		{platform.Info{OS: "linux", WSL: true}, only("cmd.exe"), "cmd.exe"},
		{platform.Info{OS: "linux"}, only("xdg-open"), "xdg-open"},
		{platform.Info{OS: "linux"}, only(), ""},
	}
	for _, c := range cases {
		got := ""
		if command := opener(c.p, c.have); command != nil {
			got = command[0]
		}
		if got != c.want {
			t.Errorf("%+v: obtido %q, esperado %q", c.p, got, c.want)
		}
	}
}
