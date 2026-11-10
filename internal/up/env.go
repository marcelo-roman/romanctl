package up

import (
	"os"
	"path/filepath"
	"strings"

	"romanctl/internal/dotenv"
)

// KV é uma variável do .env.
type KV struct{ Key, Value string }

func quote(v string) string {
	if strings.ContainsAny(v, " ;#\"'$") {
		return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return v
}

// MergeEnv só acrescenta o que falta: um valor que a pessoa mudou nunca é sobrescrito.
func MergeEnv(path string, vars []KV) ([]string, error) {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	present := map[string]bool{}
	for _, kv := range dotenv.Parse(string(existing)) {
		present[kv.Key] = true
	}
	var added []string
	text := string(existing)
	for _, kv := range vars {
		if present[kv.Key] {
			continue
		}
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += kv.Key + "=" + quote(kv.Value) + "\n"
		added = append(added, kv.Key)
	}
	if len(added) == 0 {
		return nil, nil
	}
	return added, os.WriteFile(path, []byte(text), 0o600)
}

// ProcessEnv devolve o .env para o app. O que já está no ambiente tem prioridade.
func ProcessEnv(path string, lookup func(string) (string, bool)) []string {
	return dotenv.Load(path, lookup)
}

// EnsureIgnored evita que o .env acabe em um commit.
func EnsureIgnored(projectPath string) (bool, error) {
	path := filepath.Join(projectPath, ".gitignore")
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if l := strings.TrimSpace(line); l == ".env" || l == "/.env" || l == "*.env" {
			return false, nil
		}
	}
	text := string(b)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return true, os.WriteFile(path, []byte(text+".env\n"), 0o644)
}
