// Package i18n translates romanctl's messages. The English text is the key itself, and
// pt.go holds the Portuguese version. Anything without an entry shows as written.
package i18n

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Available languages.
const (
	PT = "pt"
	EN = "en"
)

var current = detect()

// Lang returns the language in use.
func Lang() string { return current }

// Set changes the language in use.
func Set(lang string) { current = lang }

// Add adds Portuguese translations, keyed by the English text.
func Add(entries map[string]string) {
	for k, v := range entries {
		pt[k] = v
	}
}

// T translates a text.
func T(s string) string {
	if current == PT {
		if v, ok := pt[s]; ok {
			return v
		}
	}
	return s
}

func Sprintf(format string, a ...any) string { return fmt.Sprintf(T(format), a...) }
func Printf(format string, a ...any)         { fmt.Printf(T(format), a...) }
func Errorf(format string, a ...any) error   { return fmt.Errorf(T(format), a...) }

// ---------------------------------------------------------------- detection

func detect() string {
	lang, _ := Resolve(os.Args, os.Getenv, savedLang, macLanguage)
	return lang
}

// Resolve picks the language in this order: --lang, ROMANCTL_LANG, saved configuration, system
// variables and the macOS language. It also returns where the choice came from.
func Resolve(args []string, env func(string) string, saved, mac func() string) (lang, source string) {
	if isTestBinary(args) && env("ROMANCTL_LANG") == "" {
		return EN, "test"
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
	// LANGUAGE may be a list, such as pt_BR:en.
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

// Parse accepts only what the user explicitly asked for in one of the available languages.
func Parse(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pt", "pt-br", "pt_br", "portuguese", "portugues", "portugu\u00eas":
		return PT
	case "en", "en-us", "en_us", "english", "ingles", "ingl\u00eas":
		return EN
	}
	return ""
}

// parseSystem reads a system locale. C and POSIX say nothing about the language; any other
// language falls back to English, which the user is most likely to read.
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

// ---------------------------------------------------------------- saved configuration

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

// SavedLang reports the language saved with "romanctl config lang".
func SavedLang() string { return Parse(savedLang()) }

// Save stores the language, or clears the choice when lang is empty.
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

// Source explains where the language in use came from.
func Source() string {
	_, source := Resolve(os.Args, os.Getenv, savedLang, macLanguage)
	return source
}
