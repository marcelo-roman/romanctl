package i18n

import (
	"regexp"
	"testing"
)

func none() string { return "" }

func TestResolveOrdem(t *testing.T) {
	envOf := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	saved := func() string { return "pt" }
	cases := []struct {
		name         string
		args         []string
		env          map[string]string
		saved, mac   func() string
		lang, source string
	}{
		{"--lang vence tudo", []string{"romanctl", "--lang", "en"}, map[string]string{"ROMANCTL_LANG": "pt", "LANG": "pt_BR.UTF-8"}, saved, none, EN, "--lang"},
		{"--lang=", []string{"romanctl", "--lang=pt"}, nil, none, none, PT, "--lang"},
		{"variável própria", []string{"romanctl"}, map[string]string{"ROMANCTL_LANG": "en", "LANG": "pt_BR.UTF-8"}, saved, none, EN, "ROMANCTL_LANG"},
		{"salvo", []string{"romanctl"}, map[string]string{"LANG": "en_US.UTF-8"}, saved, none, PT, "config"},
		{"LC_ALL antes de LANG", []string{"romanctl"}, map[string]string{"LC_ALL": "en_GB.UTF-8", "LANG": "pt_BR.UTF-8"}, none, none, EN, "LC_ALL"},
		{"LANG português", []string{"romanctl"}, map[string]string{"LANG": "pt_BR.UTF-8"}, none, none, PT, "LANG"},
		{"Portugal", []string{"romanctl"}, map[string]string{"LANG": "pt_PT"}, none, none, PT, "LANG"},
		{"lista em LANGUAGE", []string{"romanctl"}, map[string]string{"LANGUAGE": "pt_BR:en"}, none, none, PT, "LANGUAGE"},
		{"C não decide", []string{"romanctl"}, map[string]string{"LC_ALL": "C", "LANG": "pt_BR.UTF-8"}, none, none, PT, "LANG"},
		{"C.UTF-8 não decide", []string{"romanctl"}, map[string]string{"LANG": "C.UTF-8"}, none, none, EN, "default"},
		{"outro idioma cai em inglês", []string{"romanctl"}, map[string]string{"LANG": "de_DE.UTF-8"}, none, none, EN, "LANG"},
		{"macOS", []string{"romanctl"}, nil, none, func() string { return "pt-BR" }, PT, "macOS"},
		{"sem nenhuma pista", []string{"romanctl"}, nil, none, none, EN, "default"},
		{"--lang inválido é ignorado", []string{"romanctl", "--lang", "fr"}, map[string]string{"LANG": "pt_BR"}, none, none, PT, "LANG"},
	}
	for _, c := range cases {
		lang, source := Resolve(c.args, envOf(c.env), c.saved, c.mac)
		if lang != c.lang || source != c.source {
			t.Errorf("%s: obtido %s/%s, esperado %s/%s", c.name, lang, source, c.lang, c.source)
		}
	}
}

func TestTraduz(t *testing.T) {
	defer Set(Lang())
	Set(EN)
	if T("Sim") != "Yes" {
		t.Fatal(T("Sim"))
	}
	if got := Sprintf("%s parado", "Redis"); got != "Redis stopped" {
		t.Fatal(got)
	}
	if got := T("texto que não existe no catálogo"); got != "texto que não existe no catálogo" {
		t.Fatal(got)
	}
	Set(PT)
	if T("Sim") != "Sim" {
		t.Fatal(T("Sim"))
	}
}

var verbs = regexp.MustCompile(`%[-+# 0]*[0-9.*]*[a-zA-Z%]`)

// Formato diferente entre os idiomas viraria %!s(MISSING) na tela.
func TestTraducaoTemOsMesmosVerbos(t *testing.T) {
	for pt, translated := range en {
		a, b := verbs.FindAllString(pt, -1), verbs.FindAllString(translated, -1)
		if len(a) != len(b) {
			t.Errorf("%q → %q: %v e %v", pt, translated, a, b)
			continue
		}
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("%q → %q: %v e %v", pt, translated, a, b)
				break
			}
		}
	}
}

func TestSemPortuguesNoIngles(t *testing.T) {
	accent := regexp.MustCompile(`[áàâãéêíóôõúç]`)
	for pt, translated := range en {
		if accent.MatchString(translated) {
			t.Errorf("%q → %q ainda tem acento", pt, translated)
		}
	}
}
