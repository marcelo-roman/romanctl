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

// assignmentFlags applies --global and --realms over the current ones; whatever is not on the line stays as is.
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
				return next, ui.Usagef("use --realms realm=admin|viewer, comma separated")
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
			ui.Warnf("No API client. Create one in IAM Manager → Operators → API clients and run: romanctl iam credentials set --client-id <id>")
			return ui.ErrReported
		}
		if err != nil {
			return err
		}
		return output(c, func() {
			i18n.Printf("Client: %s\nAPI:    %s\nToken:  %s\nSource: %s\n", c.ClientID, c.API, c.TokenURL, c.Source)
		})
	case "set":
		id := flagVals["client-id"]
		if id == "" {
			return ui.Usagef("usage: romanctl iam credentials set --client-id <id> [--api URL] [--token-url URL]; the secret comes from standard input")
		}
		secret, err := readSecret()
		if err != nil {
			return err
		}
		if secret == "" {
			return ui.Usagef("provide the client_secret on standard input")
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
		return output(c, func() { ui.Okf("Client %s (%s) saved to %s", c.ClientID, me.Username, c.Source) })
	case "clear":
		if err := confirmed(flags, i18n.Sprintf("Delete the credential saved in %s?", iam.CredentialsPath())); err != nil {
			return err
		}
		if err := iam.RemoveCredentials(); err != nil {
			return err
		}
		return output(map[string]any{"ok": true}, func() { ui.Okf("Credential deleted") })
	}
	return ui.Usagef("usage: romanctl iam credentials [show|set|clear]")
}

func listAPIClients(ctx context.Context) ([]iamAPIClient, error) {
	var clients []iamAPIClient
	err := iam.Do(ctx, "GET", "/api/api-clients", nil, &clients)
	return clients, err
}

// findAPIClient accepts the id, the client_id or the name.
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
	return iamAPIClient{}, i18n.Errorf("API client %s not found", who)
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
		table(i18n.T("NAME\tCLIENT\tSTATE\tGLOBAL\tPER REALM\tID"), rows)
	})
}

// saveIfCurrent: rotating the client in use invalidates the saved secret, so the new one replaces it.
func saveIfCurrent(issued iamIssuedClient) {
	current, err := iam.LoadCredentials()
	if err != nil || current.ClientID != issued.ClientID {
		return
	}
	if current.Source == "env" {
		ui.Warnf("This is the IAM_CLIENT_ID client: replace IAM_CLIENT_SECRET in the environment")
		return
	}
	current.ClientSecret = issued.Secret
	if err := iam.SaveCredentials(current); err != nil {
		ui.Warnf("Could not save the new secret in %s: %v", current.Source, err)
		return
	}
	ui.Infof("The new secret was saved in %s", current.Source)
}

const apiClientUsage = "usage: romanctl iam api-client create <name> | roles|enable|disable|rotate|delete <client>"

func iamAPIClientCommand(ctx context.Context, args []string, flags map[string]bool) error {
	if len(args) < 2 {
		return ui.Usagef(apiClientUsage)
	}
	verb, target := args[0], args[1]
	show := func(k iamIssuedClient) {
		i18n.Printf("client_id:      %s\nclient_secret:  %s\ntoken endpoint: %s\napi:            %s\n", k.ClientID, k.Secret, k.TokenEndpoint, k.APIURL)
		ui.Warnf("The secret is shown only now; keep it in a vault")
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
			ui.Okf("Client %s: %s; global [%s]; per realm [%s]", updated.Name, onOff(updated.Enabled), strings.Join(updated.Assignment.Global, ","), updated.Assignment.realmList())
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
		question := i18n.Sprintf("Delete the API client %s (%s)?", client.Name, client.ClientID)
		if current, err := iam.LoadCredentials(); err == nil && current.ClientID == client.ClientID {
			question = i18n.Sprintf("Delete the API client %s? It is the one romanctl is using.", client.Name)
		}
		if err := confirmed(flags, question); err != nil {
			return err
		}
		if err := iam.Do(ctx, "DELETE", path, nil, nil); err != nil {
			return err
		}
		return output(map[string]any{"ok": true}, func() { ui.Okf("API client %s deleted", client.Name) })
	}
	return ui.Usagef(apiClientUsage)
}
