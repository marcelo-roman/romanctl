// Package browser opens URLs in the machine's browser.
package browser

import (
	"os"
	"romanctl/internal/i18n"
	"strings"

	"romanctl/internal/platform"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

const termuxPrefix = "/data/data/com.termux/files/usr"

// opener picks the command for each environment; empty means none is available.
func opener(p platform.Info, have func(string) bool) []string {
	switch {
	case p.OS == "darwin":
		return []string{"open"}
	case have(termuxPrefix + "/bin/termux-open-url"):
		return []string{termuxPrefix + "/bin/termux-open-url"}
	case p.WSL && have("wslview"):
		return []string{"wslview"}
	case p.WSL && have("rundll32.exe"):
		// Hands the whole URL to the default Windows browser, bypassing cmd.
		return []string{"rundll32.exe", "url.dll,FileProtocolHandler"}
	case p.WSL && have("cmd.exe"):
		return []string{"cmd.exe", "/c", "start"}
	case have("xdg-open"):
		return []string{"xdg-open"}
	}
	return nil
}

// OpenURL opens the URL or, with no browser available, just prints the address.
func OpenURL(u string) error {
	command := opener(platform.Current, func(bin string) bool { return sys.Have(bin) || sys.Exists(bin) })
	if command == nil {
		ui.Warnf("Could not find a way to open the browser; open it manually: %s", u)
		return nil
	}
	var env []string
	if command[0] == termuxPrefix+"/bin/termux-open-url" {
		env = []string{"PATH=" + termuxPrefix + "/bin:" + os.Getenv("PATH"), "PREFIX=" + termuxPrefix, "LD_LIBRARY_PATH="}
	}
	target := u
	if command[0] == "cmd.exe" {
		target = cmdEscape(u)
	}
	if out, err := sys.OutputEnv(env, command[0], append(command[1:], target)...); err != nil {
		return i18n.Errorf("could not open the browser (%v) %s — open it manually: %s", err, out, u)
	}
	ui.Okf("Opening %s", ui.SBlue.Render(u))
	return nil
}

// cmdEscape protects the URL from cmd.exe, which would read & | < > ^ ( ) as syntax and cut the
// address at the first &, keeping only the first parameter.
func cmdEscape(u string) string {
	var b strings.Builder
	for _, r := range u {
		if strings.ContainsRune("^&|<>()", r) {
			b.WriteByte('^')
		}
		b.WriteRune(r)
	}
	return b.String()
}
