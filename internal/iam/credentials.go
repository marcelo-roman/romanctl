package iam

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"romanctl/internal/auth"
	"romanctl/internal/config"
	"romanctl/internal/dotenv"
	"romanctl/internal/i18n"
)

const DefaultTokenURL = "https://auth.marceloroman.com.br/realms/private/protocol/openid-connect/token"

const (
	envAPI      = "IAM_API_URL"
	envTokenURL = "IAM_TOKEN_URL"
	envClientID = "IAM_CLIENT_ID"
	envSecret   = "IAM_CLIENT_SECRET"
)

// ErrNoCredentials: sem cliente da API não há como chamar o IAM Manager.
var ErrNoCredentials = errors.New("no IAM Manager API client")

// Credentials é um cliente da API do IAM Manager (client_credentials no realm private).
type Credentials struct {
	API          string `json:"api"`
	TokenURL     string `json:"tokenUrl"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"-"`
	Source       string `json:"source"`
}

// CredentialsPath é o mesmo arquivo que a skill iam-manager usa.
func CredentialsPath() string {
	if p := os.Getenv("IAM_CREDENTIALS_FILE"); p != "" {
		return p
	}
	return filepath.Join(config.Home, ".config", "iam-manager", "credentials.env")
}

func fileValues(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, kv := range dotenv.Parse(string(data)) {
		values[kv.Key] = kv.Value
	}
	return values, nil
}

// LoadCredentials lê o arquivo e deixa as variáveis de ambiente valerem por cima, campo a campo.
func LoadCredentials() (Credentials, error) {
	path := CredentialsPath()
	values, err := fileValues(path)
	if err != nil {
		return Credentials{}, err
	}
	fromEnv := false
	value := func(key string) string {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			fromEnv = fromEnv || key == envClientID
			return v
		}
		return values[key]
	}
	c := Credentials{
		API:          strings.TrimSuffix(value(envAPI), "/"),
		TokenURL:     value(envTokenURL),
		ClientID:     value(envClientID),
		ClientSecret: value(envSecret),
		Source:       path,
	}
	if fromEnv {
		c.Source = "env"
	}
	if c.API == "" {
		c.API = auth.DefaultEndpoints().API
	}
	if c.TokenURL == "" {
		c.TokenURL = DefaultTokenURL
	}
	if c.ClientID == "" || c.ClientSecret == "" {
		return c, ErrNoCredentials
	}
	return c, nil
}

// SaveCredentials grava com 0600, trocando o arquivo de uma vez.
func SaveCredentials(c Credentials) error {
	path := CredentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	for _, v := range []string{c.API, c.TokenURL, c.ClientID, c.ClientSecret} {
		if strings.ContainsAny(v, "\n\r\"") {
			return i18n.Errorf("a credencial tem um caractere inválido")
		}
	}
	content := fmt.Sprintf("%s=%s\n%s=%s\n%s=%s\n%s=\"%s\"\n", envAPI, c.API, envTokenURL, c.TokenURL, envClientID, c.ClientID, envSecret, c.ClientSecret)
	tmp, err := os.CreateTemp(filepath.Dir(path), ".credentials-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func RemoveCredentials() error {
	err := os.Remove(CredentialsPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
