package scaffold

import (
	"bytes"
	"strings"
	"unicode"
)

// Templates are real projects that build and pass their tests in place. The sample project name is
// sample-app, the generators' sample entity is SampleItem and the --in aggregate is SampleParent.
// Generating means replacing those names, in every spelling, with the chosen ones.
var (
	sampleProject = []string{"sample", "app"}
	sampleEntity  = []string{"sample", "item"}
	sampleParent  = []string{"sample", "parent"}
)

// words splits product-item, product_item, ProductItem and my-shop.v2 into lowercase words.
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

// spellings returns a name's spellings, in the order they must be replaced.
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

// projectReplacer replaces the sample project name with the project's.
func projectReplacer(name string) *strings.Replacer {
	return strings.NewReplacer(pairs(sampleProject, words(name))...)
}

// entityReplacer also replaces the sample entity and aggregate. The plural comes first, because
// SampleItems contains SampleItem.
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

// replaceContent replaces the names in text files; binaries pass through untouched.
func replaceContent(r *strings.Replacer, b []byte) []byte {
	if bytes.IndexByte(b, 0) >= 0 {
		return b
	}
	return []byte(r.Replace(string(b)))
}
