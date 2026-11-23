package i18n

import (
	"regexp"
	"testing"
)

func none() string { return "" }

func TestResolveOrder(t *testing.T) {
	envOf := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	saved := func() string { return "pt" }
	cases := []struct {
		name         string
		args         []string
		env          map[string]string
		saved, mac   func() string
		lang, source string
	}{
		{"--lang beats everything", []string{"romanctl", "--lang", "en"}, map[string]string{"ROMANCTL_LANG": "pt", "LANG": "pt_BR.UTF-8"}, saved, none, EN, "--lang"},
		{"--lang=", []string{"romanctl", "--lang=pt"}, nil, none, none, PT, "--lang"},
		{"own variable", []string{"romanctl"}, map[string]string{"ROMANCTL_LANG": "en", "LANG": "pt_BR.UTF-8"}, saved, none, EN, "ROMANCTL_LANG"},
		{"saved", []string{"romanctl"}, map[string]string{"LANG": "en_US.UTF-8"}, saved, none, PT, "config"},
		{"LC_ALL before LANG", []string{"romanctl"}, map[string]string{"LC_ALL": "en_GB.UTF-8", "LANG": "pt_BR.UTF-8"}, none, none, EN, "LC_ALL"},
		{"Portuguese LANG", []string{"romanctl"}, map[string]string{"LANG": "pt_BR.UTF-8"}, none, none, PT, "LANG"},
		{"Portugal", []string{"romanctl"}, map[string]string{"LANG": "pt_PT"}, none, none, PT, "LANG"},
		{"list in LANGUAGE", []string{"romanctl"}, map[string]string{"LANGUAGE": "pt_BR:en"}, none, none, PT, "LANGUAGE"},
		{"C does not decide", []string{"romanctl"}, map[string]string{"LC_ALL": "C", "LANG": "pt_BR.UTF-8"}, none, none, PT, "LANG"},
		{"C.UTF-8 does not decide", []string{"romanctl"}, map[string]string{"LANG": "C.UTF-8"}, none, none, EN, "default"},
		{"another language falls back to English", []string{"romanctl"}, map[string]string{"LANG": "de_DE.UTF-8"}, none, none, EN, "LANG"},
		{"macOS", []string{"romanctl"}, nil, none, func() string { return "pt-BR" }, PT, "macOS"},
		{"no clue at all", []string{"romanctl"}, nil, none, none, EN, "default"},
		{"invalid --lang is ignored", []string{"romanctl", "--lang", "fr"}, map[string]string{"LANG": "pt_BR"}, none, none, PT, "LANG"},
	}
	for _, c := range cases {
		lang, source := Resolve(c.args, envOf(c.env), c.saved, c.mac)
		if lang != c.lang || source != c.source {
			t.Errorf("%s: got %s/%s, want %s/%s", c.name, lang, source, c.lang, c.source)
		}
	}
}

func TestTranslates(t *testing.T) {
	defer Set(Lang())
	Set(PT)
	if T("Yes") != "Sim" {
		t.Fatal(T("Yes"))
	}
	if got := Sprintf("%s stopped", "Redis"); got != "Redis parado" {
		t.Fatal(got)
	}
	if got := T("text missing from the catalog"); got != "text missing from the catalog" {
		t.Fatal(got)
	}
	Set(EN)
	if T("Yes") != "Yes" {
		t.Fatal(T("Yes"))
	}
}

var verbs = regexp.MustCompile(`%[-+# 0]*[0-9.*]*[a-zA-Z%]`)

// A format that differs between languages would show as %!s(MISSING) on screen.
func TestTranslationHasSameVerbs(t *testing.T) {
	for source, translated := range pt {
		a, b := verbs.FindAllString(source, -1), verbs.FindAllString(translated, -1)
		if len(a) != len(b) {
			t.Errorf("%q → %q: %v e %v", source, translated, a, b)
			continue
		}
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("%q → %q: %v e %v", source, translated, a, b)
				break
			}
		}
	}
}

func TestNoPortugueseInEnglish(t *testing.T) {
	accent := regexp.MustCompile(`[\x{c0}-\x{d6}\x{d8}-\x{f6}\x{f8}-\x{ff}]`)
	for source := range pt {
		if accent.MatchString(source) {
			t.Errorf("%q still has an accent", source)
		}
	}
}
