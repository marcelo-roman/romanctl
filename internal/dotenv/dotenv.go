// Package dotenv lê arquivos .env.
package dotenv

import (
	"os"
	"strings"
)

// KV é uma variável do .env.
type KV struct{ Key, Value string }

func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
		return strings.ReplaceAll(v[1:len(v)-1], `\"`, `"`)
	}
	return v
}

// Parse lê um .env; linhas em branco e comentários são ignorados.
func Parse(text string) []KV {
	var out []KV
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "export "))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, value, ok := strings.Cut(line, "="); ok && strings.TrimSpace(key) != "" {
			out = append(out, KV{strings.TrimSpace(key), unquote(value)})
		}
	}
	return out
}

// Load devolve as variáveis do arquivo no formato CHAVE=valor. O que já está no ambiente tem prioridade.
func Load(path string, lookup func(string) (string, bool)) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []string
	for _, kv := range Parse(string(b)) {
		if _, set := lookup(kv.Key); !set {
			out = append(out, kv.Key+"="+kv.Value)
		}
	}
	return out
}
