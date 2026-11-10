package scaffold

import (
	"bytes"
	"strings"
	"unicode"
)

// Os templates são projetos de verdade, que compilam e passam nos testes na própria pasta. O nome do
// projeto de exemplo é sample-app, a entidade de exemplo dos geradores é SampleItem e o agregado de
// --in é SampleParent. Gerar é trocar esses nomes, em todas as grafias, pelos escolhidos.
var (
	sampleProject = []string{"sample", "app"}
	sampleEntity  = []string{"sample", "item"}
	sampleParent  = []string{"sample", "parent"}
)

// words separa product-item, product_item, ProductItem e minha-loja.v2 em palavras minúsculas.
func words(s string) []string {
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, strings.ToLower(string(cur)))
			cur = nil
		}
	}
	runes := []rune(s)
	for i, r := range runes {
		switch {
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			flush()
		case unicode.IsUpper(r) && i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])):
			flush()
			cur = append(cur, r)
		default:
			cur = append(cur, r)
		}
	}
	flush()
	return out
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// spellings são as grafias de um nome, na ordem em que precisam ser trocadas.
func spellings(w []string) []string {
	pascal := ""
	title := make([]string, len(w))
	for i, x := range w {
		pascal += upperFirst(x)
		title[i] = upperFirst(x)
	}
	camel := ""
	if len(w) > 0 {
		camel = w[0] + pascal[len(upperFirst(w[0])):]
	}
	return []string{
		pascal,
		camel,
		strings.Join(w, "_"),
		strings.Join(w, "-"),
		strings.ToUpper(strings.Join(w, "_")),
		strings.Join(title, " "),
		strings.Join(w, " "),
		strings.Join(w, ""),
	}
}

func pairs(sample, actual []string) []string {
	from, to := spellings(sample), spellings(actual)
	out := make([]string, 0, 2*len(from))
	for i := range from {
		out = append(out, from[i], to[i])
	}
	return out
}

func pluralWords(w []string) []string {
	if len(w) == 0 {
		return w
	}
	out := append([]string(nil), w...)
	out[len(out)-1] = strings.ToLower(Plural(out[len(out)-1]))
	return out
}

// projectReplacer troca o nome do projeto de exemplo pelo do projeto.
func projectReplacer(name string) *strings.Replacer {
	return strings.NewReplacer(pairs(sampleProject, words(name))...)
}

// entityReplacer troca também a entidade e o agregado de exemplo. O plural vem antes, porque
// SampleItems contém SampleItem.
func entityReplacer(project, entity, plural, parent string) *strings.Replacer {
	var all []string
	all = append(all, pairs(pluralWords(sampleEntity), words(plural))...)
	all = append(all, pairs(sampleEntity, words(entity))...)
	if parent != "" {
		all = append(all, pairs(pluralWords(sampleParent), pluralWords(words(parent)))...)
		all = append(all, pairs(sampleParent, words(parent))...)
	}
	all = append(all, pairs(sampleProject, words(project))...)
	return strings.NewReplacer(all...)
}

// replaceContent troca os nomes em arquivos de texto; binários passam intactos.
func replaceContent(r *strings.Replacer, b []byte) []byte {
	if bytes.IndexByte(b, 0) >= 0 {
		return b
	}
	return []byte(r.Replace(string(b)))
}
