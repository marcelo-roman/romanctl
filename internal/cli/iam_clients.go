package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"

	"romanctl/internal/auth"
	"romanctl/internal/i18n"
	"romanctl/internal/iam"
	"romanctl/internal/ui"
)

type iamAssignment struct {
	Global    []string          `json:"global"`
	Inherited []string          `json:"inherited,omitempty"`
	Realms    map[string]string `json:"realms"`
}

type iamAPIClient struct {
	ID          string        `json:"id"`
	ClientID    string        `json:"clientId"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Enabled     bool          `json:"enabled"`
	CreatedBy   string        `json:"createdBy,omitempty"`
	Assignment  iamAssignment `json:"assignment"`
}

type iamIssuedClient struct {
	iamAPIClient
	Secret        string `json:"secret"`
	TokenEndpoint string `json:"tokenEndpoint"`
	APIURL        string `json:"apiUrl"`
}

func (a iamAssignment) writable() iamAssignment {
	global, realms := a.Global, a.Realms
	if global == nil {
		global = []string{}
	}
	if realms == nil {
		realms = map[string]string{}
	}
	return iamAssignment{Global: global, Realms: realms}
}

func (a iamAssignment) realmList() string {
	var out []string
	for realm, role := range a.Realms {
		out = append(out, realm+"="+role)
	}
	return strings.Join(out, ",")
}

// assignmentFlags aplica --global e --realms sobre o atual; o que não vier na linha fica como está.
func assignmentFlags(current iamAssignment) (iamAssignment, error) {
	next := current.writable()
	if v, ok := flagVals["global"]; ok {
		next.Global = list(v)
		if next.Global == nil {
			next.Global = []string{}
		}
	}
	if v, ok := flagVals["realms"]; ok {
		next.Realms = map[string]string{}
		for _, pair := range list(v) {
			name, role, ok := strings.Cut(pair, "=")
			if !ok {
				return next, ui.Usagef("use --realms realm=admin|viewer, separados por vírgula")
			}
			next.Realms[name] = role
		}
	}
	return next, nil
}

func readSecret() (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, 4096))
		return strings.TrimSpace(string(data)), err
	}
	var secret string
	field := huh.NewInput().Title("client_secret").EchoMode(huh.EchoModePassword).Value(&secret)
	if err := ui.Ask(false, field); err != nil {
		return "", err
	}
	return strings.TrimSpace(secret), nil
}

func iamCredentials(ctx context.Context, args []string, flags map[string]bool) error {
	switch arg(args, 0) {
	case "", "show":
		c, err := iam.LoadCredentials()
		if errors.Is(err, iam.ErrNoCredentials) {
			if ui.JSONOutput {
				return ui.PrintJSON(map[string]any{"configured": false, "path": iam.CredentialsPath()})
			}
			ui.Warnf("Sem cliente da API. Crie um em IAM Manager → Operadores → Clientes da API e rode: romanctl iam credentials set --client-id <id>")
			return ui.ErrReported
		}
		if err != nil {
			return err
		}
		return output(c, func() {
			i18n.Printf("Cliente: %s\nAPI:     %s\nToken:   %s\nOrigem:  %s\n", c.ClientID, c.API, c.TokenURL, c.Source)
		})
	case "set":
		id := flagVals["client-id"]
		if id == "" {
			return ui.Usagef("use: romanctl iam credentials set --client-id <id> [--api URL] [--token-url URL]; o secret vem pela entrada padrão")
		}
		secret, err := readSecret()
		if err != nil {
			return err
		}
		if secret == "" {
			return ui.Usagef("informe o client_secret pela entrada padrão")
		}
		c := iam.Credentials{
			API:          strings.TrimSuffix(orDefault(flagVals["api"], auth.DefaultEndpoints().API), "/"),
			TokenURL:     orDefault(flagVals["token-url"], iam.DefaultTokenURL),
			ClientID:     id,
			ClientSecret: secret,
		}
		raw, err := iam.With(ctx, c, "GET", "/api/me", nil)
		if err != nil {
			return err
		}
		var me struct {
			Username string `json:"username"`
		}
		_ = json.Unmarshal(raw, &me)
		if err := iam.SaveCredentials(c); err != nil {
			return err
		}
		c.Source = iam.CredentialsPath()
		return output(c, func() { ui.Okf("Cliente %s (%s) salvo em %s", c.ClientID, me.Username, c.Source) })
	case "clear":
		if err := confirmed(flags, i18n.Sprintf("Apagar a credencial salva em %s?", iam.CredentialsPath())); err != nil {
			return err
		}
		if err := iam.RemoveCredentials(); err != nil {
			return err
		}
		return output(map[string]any{"ok": true}, func() { ui.Okf("Credencial apagada") })
	}
	return ui.Usagef("use: romanctl iam credentials [show|set|clear]")
}

func listAPIClients(ctx context.Context) ([]iamAPIClient, error) {
	var clients []iamAPIClient
	err := iam.Do(ctx, "GET", "/api/api-clients", nil, &clients)
	return clients, err
}

// findAPIClient aceita o id, o client_id ou o nome.
func findAPIClient(ctx context.Context, who string) (iamAPIClient, error) {
	clients, err := listAPIClients(ctx)
	if err != nil {
		return iamAPIClient{}, err
	}
	for _, c := range clients {
		if c.ID == who || c.ClientID == who || strings.EqualFold(c.Name, who) {
			return c, nil
		}
	}
	return iamAPIClient{}, i18n.Errorf("cliente da API %s não encontrado", who)
}

func iamAPIClients(ctx context.Context) error {
	clients, err := listAPIClients(ctx)
	if err != nil {
		return err
	}
	return output(clients, func() {
		var rows [][]string
		for _, c := range clients {
			global := append(append([]string{}, c.Assignment.Global...), c.Assignment.Inherited...)
			rows = append(rows, []string{c.Name, c.ClientID, onOff(c.Enabled), strings.Join(global, ","), c.Assignment.realmList(), c.ID})
		}
		table(i18n.T("NOME\tCLIENT\tESTADO\tGLOBAIS\tPOR REALM\tID"), rows)
	})
}

// saveIfCurrent: rotacionar o próprio cliente derruba o secret salvo; o novo entra no lugar.
func saveIfCurrent(issued iamIssuedClient) {
	current, err := iam.LoadCredentials()
	if err != nil || current.ClientID != issued.ClientID {
		return
	}
	if current.Source == "env" {
		ui.Warnf("Este é o cliente do IAM_CLIENT_ID: troque o IAM_CLIENT_SECRET no ambiente")
		return
	}
	current.ClientSecret = issued.Secret
	if err := iam.SaveCredentials(current); err != nil {
		ui.Warnf("Não consegui salvar o novo secret em %s: %v", current.Source, err)
		return
	}
	ui.Infof("O secret novo foi salvo em %s", current.Source)
}

const apiClientUsage = "use: romanctl iam api-client create <nome> | roles|enable|disable|rotate|delete <cliente>"

func iamAPIClientCommand(ctx context.Context, args []string, flags map[string]bool) error {
	if len(args) < 2 {
		return ui.Usagef(apiClientUsage)
	}
	verb, target := args[0], args[1]
	show := func(k iamIssuedClient) {
		i18n.Printf("client_id:      %s\nclient_secret:  %s\ntoken endpoint: %s\napi:            %s\n", k.ClientID, k.Secret, k.TokenEndpoint, k.APIURL)
		ui.Warnf("O secret só aparece agora; guarde num cofre")
	}
	if verb == "create" {
		assignment, err := assignmentFlags(iamAssignment{})
		if err != nil {
			return err
		}
		body := map[string]any{"name": target, "description": flagVals["description"], "assignment": assignment}
		var issued iamIssuedClient
		if err := iam.Do(ctx, "POST", "/api/api-clients", body, &issued); err != nil {
			return err
		}
		return output(issued, func() { show(issued) })
	}
	client, err := findAPIClient(ctx, target)
	if err != nil {
		return err
	}
	path := "/api/api-clients/" + url.PathEscape(client.ID)
	switch verb {
	case "roles", "enable", "disable":
		assignment, err := assignmentFlags(client.Assignment)
		if err != nil {
			return err
		}
		enabled := client.Enabled
		if verb != "roles" {
			enabled = verb == "enable"
		}
		var updated iamAPIClient
		if err := iam.Do(ctx, "PUT", path, map[string]any{"enabled": enabled, "assignment": assignment}, &updated); err != nil {
			return err
		}
		return output(updated, func() {
			ui.Okf("Cliente %s: %s; globais [%s]; por realm [%s]", updated.Name, onOff(updated.Enabled), strings.Join(updated.Assignment.Global, ","), updated.Assignment.realmList())
		})
	case "rotate":
		var issued iamIssuedClient
		if err := iam.Do(ctx, "POST", path+"/rotate", nil, &issued); err != nil {
			return err
		}
		if err := output(issued, func() { show(issued) }); err != nil {
			return err
		}
		saveIfCurrent(issued)
		return nil
	case "delete":
		question := i18n.Sprintf("Excluir o cliente da API %s (%s)?", client.Name, client.ClientID)
		if current, err := iam.LoadCredentials(); err == nil && current.ClientID == client.ClientID {
			question = i18n.Sprintf("Excluir o cliente da API %s? É o que o romanctl está usando.", client.Name)
		}
		if err := confirmed(flags, question); err != nil {
			return err
		}
		if err := iam.Do(ctx, "DELETE", path, nil, nil); err != nil {
			return err
		}
		return output(map[string]any{"ok": true}, func() { ui.Okf("Cliente da API %s excluído", client.Name) })
	}
	return ui.Usagef(apiClientUsage)
}
