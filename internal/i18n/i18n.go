// Package i18n traduz as mensagens do romanctl. O texto em português é a própria chave,
// e en.go diz a versão em inglês. O que não tem entrada aparece como foi escrito.
package i18n

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Idiomas disponíveis.
const (
	PT = "pt"
	EN = "en"
)

var current = detect()

// Lang devolve o idioma em uso.
func Lang() string { return current }

// Set troca o idioma em uso.
func Set(lang string) { current = lang }

// Add acrescenta traduções, para textos longos que ficam junto do código que os usa.
func Add(entries map[string]string) {
	for k, v := range entries {
		en[k] = v
	}
}

// T traduz um texto.
func T(s string) string {
	if current == EN {
		if v, ok := en[s]; ok {
			return v
		}
	}
	return s
}

func Sprintf(format string, a ...any) string { return fmt.Sprintf(T(format), a...) }
func Printf(format string, a ...any)         { fmt.Printf(T(format), a...) }
func Errorf(format string, a ...any) error   { return fmt.Errorf(T(format), a...) }

// ---------------------------------------------------------------- detecção

func detect() string {
	lang, _ := Resolve(os.Args, os.Getenv, savedLang, macLanguage)
	return lang
}

// Resolve escolhe o idioma, nesta ordem: --lang, ROMANCTL_LANG, configuração salva, variáveis
// do sistema e idioma do macOS. Devolve também de onde veio a escolha.
func Resolve(args []string, env func(string) string, saved, mac func() string) (lang, source string) {
	if isTestBinary(args) && env("ROMANCTL_LANG") == "" {
		return PT, "test"
	}
	for i, a := range args {
		if a == "--lang" && i+1 < len(args) {
			if l := Parse(args[i+1]); l != "" {
				return l, "--lang"
			}
		}
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			if l := Parse(v); l != "" {
				return l, "--lang"
			}
		}
	}
	if l := Parse(env("ROMANCTL_LANG")); l != "" {
		return l, "ROMANCTL_LANG"
	}
	if l := Parse(saved()); l != "" {
		return l, "config"
	}
	// LANGUAGE pode ser uma lista, como pt_BR:en.
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"} {
		first, _, _ := strings.Cut(env(name), ":")
		if l := parseSystem(first); l != "" {
			return l, name
		}
	}
	if l := parseSystem(mac()); l != "" {
		return l, "macOS"
	}
	return EN, "default"
}

func isTestBinary(args []string) bool {
	return len(args) > 0 && strings.HasSuffix(args[0], ".test")
}

// Parse aceita só o que o usuário pediu explicitamente em uma das línguas disponíveis.
func Parse(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pt", "pt-br", "pt_br", "portugues", "português":
		return PT
	case "en", "en-us", "en_us", "english", "ingles", "inglês":
		return EN
	}
	return ""
}

// parseSystem lê um locale do sistema. C e POSIX não dizem nada sobre o idioma; qualquer
// outro idioma cai em inglês, que é o que a pessoa tem mais chance de ler.
func parseSystem(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	base, _, _ := strings.Cut(locale, ".")
	if base == "" || base == "c" || base == "posix" {
		return ""
	}
	if base == "pt" || strings.HasPrefix(base, "pt_") || strings.HasPrefix(base, "pt-") {
		return PT
	}
	return EN
}

var macLang = regexp.MustCompile(`[a-zA-Z]{2}[-_][a-zA-Z-]+|\b[a-z]{2}\b`)

func macLanguage() string {
	if _, err := os.Stat("/System/Library/CoreServices/SystemVersion.plist"); err != nil {
		return ""
	}
	out, err := exec.Command("defaults", "read", "-g", "AppleLanguages").Output()
	if err != nil {
		return ""
	}
	return macLang.FindString(string(out))
}

// ---------------------------------------------------------------- configuração salva

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "romanctl", "config.yaml")
}

func savedLang() string {
	b, err := os.ReadFile(configPath())
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "lang:"); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// SavedLang informa o idioma salvo com "romanctl config lang".
func SavedLang() string { return Parse(savedLang()) }

// Save grava o idioma, ou apaga a escolha quando lang é vazio.
func Save(lang string) error {
	if lang == "" {
		if err := os.Remove(configPath()); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(configPath()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(configPath(), []byte("lang: "+lang+"\n"), 0o644)
}

// Source explica de onde veio o idioma em uso.
func Source() string {
	_, source := Resolve(os.Args, os.Getenv, savedLang, macLanguage)
	return source
}
