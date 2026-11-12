package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/iam"
	"romanctl/internal/ui"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type iamUser struct {
	ID              string   `json:"id"`
	Username        string   `json:"username"`
	Email           string   `json:"email,omitempty"`
	FirstName       string   `json:"firstName,omitempty"`
	LastName        string   `json:"lastName,omitempty"`
	Enabled         bool     `json:"enabled"`
	EmailVerified   bool     `json:"emailVerified"`
	RequiredActions []string `json:"requiredActions,omitempty"`
}

type iamCredential struct {
	Mode              string   `json:"mode"`
	Link              string   `json:"link,omitempty"`
	TemporaryPassword string   `json:"temporaryPassword,omitempty"`
	ExpiresAt         string   `json:"expiresAt,omitempty"`
	Emailed           []string `json:"emailed,omitempty"`
}

func table(header string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, header)
	for _, r := range rows {
		fmt.Fprintln(w, strings.Join(r, "\t"))
	}
	w.Flush()
}

func onOff(v bool) string {
	if v {
		return i18n.T("ativo")
	}
	return i18n.T("desligado")
}

func list(v string) []string {
	var out []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func confirmed(flags map[string]bool, question string) error {
	if flags["y"] || flags["yes"] {
		return nil
	}
	if !ui.Interactive() {
		return ui.Usagef("confirme com --yes")
	}
	ok, err := ui.Confirm(question)
	if err != nil {
		return err
	}
	if !ok {
		return ui.ErrCanceled
	}
	return nil
}

// output imprime o JSON da API com --json; sem ele, chama o formato de terminal.
func output(v any, human func()) error {
	if ui.JSONOutput {
		return ui.PrintJSON(v)
	}
	human()
	return nil
}

func iamCommand(args []string, flags map[string]bool) error {
	ctx := context.Background()
	need := func(n int, usage string) error {
		if len(args) < n {
			return ui.Usagef("use: romanctl iam %s", usage)
		}
		return nil
	}
	switch arg(args, 0) {
	case "me":
		var me struct {
			Username          string              `json:"username"`
			Realm             string              `json:"realm"`
			Roles             []string            `json:"roles"`
			GlobalPermissions []string            `json:"globalPermissions"`
			RealmPermissions  map[string][]string `json:"realmPermissions"`
		}
		if err := iam.Do(ctx, "GET", "/api/me", nil, &me); err != nil {
			return err
		}
		return output(me, func() {
			i18n.Printf("%s (realm %s)\n", me.Username, me.Realm)
			if len(me.GlobalPermissions) == 0 && len(me.RealmPermissions) == 0 {
				ui.Warnf("O cliente %s não tem permissão no IAM Manager. Dê papéis a ele em Operadores → Clientes da API", me.Username)
				return
			}
			if len(me.GlobalPermissions) > 0 {
				i18n.Printf("Em todos os realms: %s\n", strings.Join(me.GlobalPermissions, ", "))
			}
			for realm, perms := range me.RealmPermissions {
				i18n.Printf("Em %s: %s\n", realm, strings.Join(perms, ", "))
			}
		})
	case "realms":
		var realms []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			Enabled     bool   `json:"enabled"`
		}
		if err := iam.Do(ctx, "GET", "/api/realms", nil, &realms); err != nil {
			return err
		}
		return output(realms, func() {
			var rows [][]string
			for _, r := range realms {
				rows = append(rows, []string{r.Name, r.DisplayName, onOff(r.Enabled)})
			}
			table(i18n.T("REALM\tNOME\tESTADO"), rows)
		})
	case "realm":
		return iamRealm(ctx, args[1:])
	case "users":
		if err := need(2, "users <realm> [--search X]"); err != nil {
			return err
		}
		q := url.Values{"max": {"100"}}
		if s := flagVals["search"]; s != "" {
			q.Set("search", s)
		}
		var page struct {
			Users []iamUser `json:"users"`
			Total int       `json:"total"`
		}
		if err := iam.Do(ctx, "GET", "/api/realms/"+url.PathEscape(args[1])+"/users?"+q.Encode(), nil, &page); err != nil {
			return err
		}
		return output(page, func() {
			var rows [][]string
			for _, u := range page.Users {
				rows = append(rows, []string{u.Username, u.Email, strings.TrimSpace(u.FirstName + " " + u.LastName), onOff(u.Enabled), u.ID})
			}
			table(i18n.T("USUÁRIO\tE-MAIL\tNOME\tESTADO\tID"), rows)
			if page.Total > len(page.Users) {
				i18n.Printf("%d de %d; refine com --search\n", len(page.Users), page.Total)
			}
		})
	case "user":
		return iamUserCommand(ctx, args[1:], flags)
	case "api-keys":
		if err := need(2, "api-keys <realm>"); err != nil {
			return err
		}
		var keys []struct {
			ID          string   `json:"id"`
			ClientID    string   `json:"clientId"`
			Name        string   `json:"name"`
			Enabled     bool     `json:"enabled"`
			Roles       []string `json:"roles"`
			CreatedBy   string   `json:"createdBy"`
			Description string   `json:"description"`
		}
		if err := iam.Do(ctx, "GET", "/api/realms/"+url.PathEscape(args[1])+"/api-keys", nil, &keys); err != nil {
			return err
		}
		return output(keys, func() {
			var rows [][]string
			for _, k := range keys {
				rows = append(rows, []string{k.Name, k.ClientID, onOff(k.Enabled), strings.Join(k.Roles, ","), k.ID})
			}
			table(i18n.T("NOME\tCLIENT\tESTADO\tPAPÉIS\tID"), rows)
		})
	case "api-key":
		return iamAPIKey(ctx, args[1:], flags)
	case "credentials", "credential":
		return iamCredentials(ctx, args[1:], flags)
	case "api-clients", "clients":
		return iamAPIClients(ctx)
	case "api-client", "client":
		return iamAPIClientCommand(ctx, args[1:], flags)
	case "operators":
		var operators []iamOperator
		if err := iam.Do(ctx, "GET", "/api/operators", nil, &operators); err != nil {
			return err
		}
		return output(operators, func() {
			var rows [][]string
			for _, o := range operators {
				global := append(append([]string{}, o.Assignment.Global...), o.Assignment.Inherited...)
				rows = append(rows, []string{o.User.Username, o.User.Email, strings.Join(global, ","), o.Assignment.realmList(), onOff(o.User.Enabled)})
			}
			table(i18n.T("USUÁRIO\tE-MAIL\tGLOBAIS\tPOR REALM\tESTADO"), rows)
		})
	case "operator":
		return iamOperatorCommand(ctx, args[1:])
	case "apps":
		if err := need(2, "apps <realm>"); err != nil {
			return err
		}
		apps, err := listApplications(ctx, args[1])
		if err != nil {
			return err
		}
		return output(apps, func() {
			var rows [][]string
			for _, a := range apps {
				var clients []string
				for _, c := range a.Clients {
					clients = append(clients, c.ClientID)
				}
				rows = append(rows, []string{a.ID, a.Name, strings.Join(a.Roles, ","), a.Theme, strings.Join(clients, ",")})
			}
			table(i18n.T("APLICAÇÃO\tNOME\tROLES\tTEMA\tCLIENTS"), rows)
		})
	case "app":
		return iamApplicationCommand(ctx, args[1:])
	case "events":
		if err := need(2, "events <realm> [--limit N]"); err != nil {
			return err
		}
		var events struct {
			User []struct {
				Time      int64  `json:"time"`
				Type      string `json:"type"`
				UserID    string `json:"userId"`
				ClientID  string `json:"clientId"`
				IPAddress string `json:"ipAddress"`
				Error     string `json:"error"`
			} `json:"user"`
			Admin []struct {
				Time          int64  `json:"time"`
				OperationType string `json:"operationType"`
				ResourceType  string `json:"resourceType"`
				ResourcePath  string `json:"resourcePath"`
			} `json:"admin"`
		}
		if err := iam.Do(ctx, "GET", "/api/realms/"+url.PathEscape(args[1])+"/events?max="+limit("100"), nil, &events); err != nil {
			return err
		}
		return output(events, func() {
			var rows [][]string
			for _, e := range events.User {
				rows = append(rows, []string{when(e.Time), e.Type, e.ClientID, e.IPAddress, e.Error})
			}
			table(i18n.T("QUANDO\tEVENTO\tCLIENT\tIP\tERRO"), rows)
			if len(events.Admin) > 0 {
				fmt.Println()
				rows = nil
				for _, e := range events.Admin {
					rows = append(rows, []string{when(e.Time), e.OperationType, e.ResourceType, e.ResourcePath})
				}
				table(i18n.T("QUANDO\tOPERAÇÃO\tRECURSO\tCAMINHO"), rows)
			}
		})
	case "audit":
		var events []struct {
			At     string `json:"at"`
			Actor  string `json:"actor"`
			Action string `json:"action"`
			Realm  string `json:"realm"`
			Target string `json:"target"`
		}
		if err := iam.Do(ctx, "GET", "/api/audit?limit="+limit("100"), nil, &events); err != nil {
			return err
		}
		return output(events, func() {
			var rows [][]string
			for _, e := range events {
				rows = append(rows, []string{e.At, e.Actor, e.Action, e.Realm, e.Target})
			}
			table(i18n.T("QUANDO\tQUEM\tAÇÃO\tREALM\tALVO"), rows)
		})
	case "api":
		if err := need(3, "api <GET|POST|PUT|DELETE> <caminho> [--data JSON]"); err != nil {
			return err
		}
		method := strings.ToUpper(args[1])
		var body any
		if d := flagVals["data"]; d != "" {
			if !json.Valid([]byte(d)) {
				return ui.Usagef("--data precisa ser JSON válido")
			}
			body = json.RawMessage(d)
		}
		raw, err := iam.Raw(ctx, method, args[2], body)
		if err != nil {
			return err
		}
		if len(raw) == 0 {
			if ui.JSONOutput {
				return ui.PrintJSON(map[string]any{"ok": true})
			}
			ui.Okf("Feito")
			return nil
		}
		var pretty any
		if json.Unmarshal(raw, &pretty) == nil {
			return ui.PrintJSON(pretty)
		}
		fmt.Println(string(raw))
		return nil
	}
	return ui.Usagef("use: romanctl iam <comando> (veja: romanctl help iam)")
}

type iamOperator struct {
	User       iamUser       `json:"user"`
	Assignment iamAssignment `json:"assignment"`
}

// findOperator aceita o nome de usuário, o e-mail ou o id.
func findOperator(ctx context.Context, who string) (iamOperator, error) {
	var operators []iamOperator
	if err := iam.Do(ctx, "GET", "/api/operators?"+url.Values{"q": {who}}.Encode(), nil, &operators); err != nil {
		return iamOperator{}, err
	}
	for _, o := range operators {
		if o.User.ID == who || strings.EqualFold(o.User.Username, who) || strings.EqualFold(o.User.Email, who) {
			return o, nil
		}
	}
	return iamOperator{}, i18n.Errorf("operador %s não encontrado", who)
}

const operatorUsage = "use: romanctl iam operator add <e-mail> [--first-name X] [--last-name Y] | roles <operador> [--global iam-admin,...] [--realms public=admin]"

func iamOperatorCommand(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return ui.Usagef(operatorUsage)
	}
	verb, who := args[0], args[1]
	switch verb {
	case "add":
		assignment, err := assignmentFlags(iamAssignment{})
		if err != nil {
			return err
		}
		body := map[string]any{"email": who, "assignment": assignment}
		for field, flag := range map[string]string{"firstName": "first-name", "lastName": "last-name"} {
			if v := flagVals[flag]; v != "" {
				body[field] = v
			}
		}
		var created iamOperator
		if err := iam.Do(ctx, "POST", "/api/operators", body, &created); err != nil {
			return err
		}
		return output(created, func() { ui.Okf("Operador %s cadastrado; entra com a conta do Google Workspace", who) })
	case "roles":
		operator, err := findOperator(ctx, who)
		if err != nil {
			return err
		}
		assignment, err := assignmentFlags(operator.Assignment)
		if err != nil {
			return err
		}
		var updated iamOperator
		if err := iam.Do(ctx, "PUT", "/api/operators/"+url.PathEscape(operator.User.ID)+"/roles", assignment, &updated); err != nil {
			return err
		}
		return output(updated, func() { ui.Okf("Papéis de %s atualizados", who) })
	}
	return ui.Usagef(operatorUsage)
}

type iamApplication struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
	Theme    string   `json:"theme,omitempty"`
	Declared bool     `json:"declared"`
	Clients  []struct {
		ClientID string `json:"clientId"`
		Name     string `json:"name"`
		Public   bool   `json:"public"`
		Enabled  bool   `json:"enabled"`
		Declared bool   `json:"declared"`
	} `json:"clients"`
}

func listApplications(ctx context.Context, realm string) ([]iamApplication, error) {
	var apps []iamApplication
	err := iam.Do(ctx, "GET", "/api/realms/"+url.PathEscape(realm)+"/applications", nil, &apps)
	return apps, err
}

const applicationUsage = "use: romanctl iam app create <realm> <id> --roles a,b [--name X] | role <realm> <app> <role> | theme <realm> <app> [tema]"

func iamApplicationCommand(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return ui.Usagef(applicationUsage)
	}
	verb, realm, id := args[0], args[1], args[2]
	base := "/api/realms/" + url.PathEscape(realm) + "/applications"
	var app iamApplication
	switch verb {
	case "create":
		roles := list(flagVals["roles"])
		if len(roles) == 0 {
			return ui.Usagef("informe ao menos uma role com --roles a,b")
		}
		body := map[string]any{"id": id, "name": orDefault(flagVals["name"], id), "roles": roles}
		if err := iam.Do(ctx, "POST", base, body, &app); err != nil {
			return err
		}
		return output(app, func() {
			ui.Okf("Aplicação %s criada em %s com as roles %s", app.ID, realm, strings.Join(app.Roles, ", "))
		})
	case "role":
		if len(args) < 4 {
			return ui.Usagef(applicationUsage)
		}
		if err := iam.Do(ctx, "POST", base+"/"+url.PathEscape(id)+"/roles", map[string]string{"name": args[3]}, &app); err != nil {
			return err
		}
		return output(app, func() { ui.Okf("Aplicação %s: roles %s", app.ID, strings.Join(app.Roles, ", ")) })
	case "theme":
		if err := iam.Do(ctx, "PUT", base+"/"+url.PathEscape(id)+"/theme", map[string]string{"theme": arg(args, 3)}, &app); err != nil {
			return err
		}
		return output(app, func() {
			if app.Theme == "" {
				ui.Okf("Aplicação %s usa o tema do realm", app.ID)
				return
			}
			ui.Okf("Aplicação %s usa o tema %s", app.ID, app.Theme)
		})
	}
	return ui.Usagef(applicationUsage)
}

func limit(def string) string {
	if v := flagVals["limit"]; v != "" {
		if _, err := strconv.Atoi(v); err == nil {
			return v
		}
	}
	return def
}

func when(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).Local().Format("02/01 15:04:05")
}

func iamRealm(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return ui.Usagef("use: romanctl iam realm <realm>")
	}
	var settings map[string]any
	if err := iam.Do(ctx, "GET", "/api/realms/"+url.PathEscape(args[0])+"/settings", nil, &settings); err != nil {
		return err
	}
	return ui.PrintJSON(settings)
}

// userID aceita o id ou o nome de usuário, que é procurado com correspondência exata.
func userID(ctx context.Context, realm, who string) (string, error) {
	if uuidPattern.MatchString(who) {
		return who, nil
	}
	var page struct {
		Users []iamUser `json:"users"`
	}
	q := url.Values{"search": {who}, "max": {"100"}}
	if err := iam.Do(ctx, "GET", "/api/realms/"+url.PathEscape(realm)+"/users?"+q.Encode(), nil, &page); err != nil {
		return "", err
	}
	for _, u := range page.Users {
		if strings.EqualFold(u.Username, who) {
			return u.ID, nil
		}
	}
	return "", i18n.Errorf("usuário %s não encontrado em %s", who, realm)
}

func printCredential(c *iamCredential) {
	if c == nil || c.Mode == "" || c.Mode == "none" {
		return
	}
	switch {
	case c.Link != "":
		i18n.Printf("Link para definir a senha (uso único, até %s):\n  %s\n", c.ExpiresAt, c.Link)
	case c.TemporaryPassword != "":
		i18n.Printf("Senha temporária (troca no primeiro login): %s\n", c.TemporaryPassword)
	case len(c.Emailed) > 0:
		i18n.Printf("E-mail enviado pelo Keycloak para %s\n", strings.Join(c.Emailed, ", "))
	}
}

func iamUserCommand(ctx context.Context, args []string, flags map[string]bool) error {
	if len(args) < 3 {
		return ui.Usagef("use: romanctl iam user show|create|enable|disable|roles|access|reset|logout|delete <realm> <usuário>")
	}
	verb, realm, who := args[0], args[1], args[2]
	base := "/api/realms/" + url.PathEscape(realm) + "/users"
	if verb == "create" {
		body := map[string]any{
			"profile": map[string]any{"username": who, "email": flagVals["email"], "firstName": flagVals["first-name"],
				"lastName": flagVals["last-name"], "enabled": true, "emailVerified": flagVals["email"] != ""},
			"credential": orDefault(flagVals["credential"], "link"),
			"roles":      list(flagVals["roles"]),
			"groups":     []string{},
		}
		var created struct {
			User       iamUser        `json:"user"`
			Credential *iamCredential `json:"credential"`
		}
		if err := iam.Do(ctx, "POST", base, body, &created); err != nil {
			return err
		}
		return output(created, func() {
			ui.Okf("Usuário %s criado em %s", created.User.Username, realm)
			printCredential(created.Credential)
		})
	}
	id, err := userID(ctx, realm, who)
	if err != nil {
		return err
	}
	path := base + "/" + url.PathEscape(id)
	switch verb {
	case "show":
		var detail map[string]any
		if err := iam.Do(ctx, "GET", path, nil, &detail); err != nil {
			return err
		}
		return ui.PrintJSON(detail)
	case "enable", "disable":
		var u iamUser
		if err := iam.Do(ctx, "GET", path, nil, &u); err != nil {
			return err
		}
		u.Enabled = verb == "enable"
		profile := map[string]any{"username": u.Username, "email": u.Email, "firstName": u.FirstName, "lastName": u.LastName,
			"enabled": u.Enabled, "emailVerified": u.EmailVerified}
		var updated json.RawMessage
		if err := iam.Do(ctx, "PUT", path, profile, &updated); err != nil {
			return err
		}
		return output(updated, func() {
			if u.Enabled {
				ui.Okf("Usuário %s ativado", who)
			} else {
				ui.Okf("Usuário %s bloqueado", who)
			}
		})
	case "roles":
		if _, ok := flagVals["roles"]; !ok {
			return ui.Usagef("informe --roles a,b (vazio tira todos)")
		}
		var detail struct {
			Groups []struct {
				ID string `json:"id"`
			} `json:"groups"`
		}
		if err := iam.Do(ctx, "GET", path, nil, &detail); err != nil {
			return err
		}
		groups := []string{}
		for _, g := range detail.Groups {
			groups = append(groups, g.ID)
		}
		roles := list(flagVals["roles"])
		if roles == nil {
			roles = []string{}
		}
		var updated json.RawMessage
		if err := iam.Do(ctx, "PUT", path+"/membership", map[string]any{"roles": roles, "groups": groups}, &updated); err != nil {
			return err
		}
		return output(updated, func() { ui.Okf("Papéis de %s: %s", who, strings.Join(roles, ", ")) })
	case "access":
		return userAccess(ctx, path, who, arg(args, 3))
	case "reset":
		var c iamCredential
		if err := iam.Do(ctx, "POST", path+"/credential", map[string]string{"mode": orDefault(flagVals["credential"], "link")}, &c); err != nil {
			return err
		}
		return output(c, func() { printCredential(&c) })
	case "logout":
		if err := iam.Do(ctx, "POST", path+"/logout", nil, nil); err != nil {
			return err
		}
		return output(map[string]any{"ok": true}, func() { ui.Okf("Sessões de %s encerradas", who) })
	case "delete":
		if err := confirmed(flags, i18n.Sprintf("Excluir %s de %s?", who, realm)); err != nil {
			return err
		}
		if err := iam.Do(ctx, "DELETE", path, nil, nil); err != nil {
			return err
		}
		return output(map[string]any{"ok": true}, func() { ui.Okf("Usuário %s excluído", who) })
	}
	return ui.Usagef("use: romanctl iam user show|create|enable|disable|roles|access|reset|logout|delete <realm> <usuário>")
}

type iamAccess struct {
	Application string   `json:"application"`
	Name        string   `json:"name"`
	Available   []string `json:"available"`
	Roles       []string `json:"roles"`
	Effective   []string `json:"effective"`
}

// userAccess mostra as roles do usuário em cada aplicação; com app, define as dessa aplicação.
func userAccess(ctx context.Context, path, who, app string) error {
	var access []iamAccess
	if app != "" {
		if _, ok := flagVals["roles"]; !ok {
			return ui.Usagef("informe --roles a,b (vazio revoga o acesso)")
		}
		roles := list(flagVals["roles"])
		if roles == nil {
			roles = []string{}
		}
		if err := iam.Do(ctx, "PUT", path+"/access/"+url.PathEscape(app), map[string]any{"roles": roles}, &access); err != nil {
			return err
		}
	} else if err := iam.Do(ctx, "GET", path+"/access", nil, &access); err != nil {
		return err
	}
	return output(access, func() {
		var rows [][]string
		for _, a := range access {
			rows = append(rows, []string{a.Application, strings.Join(a.Roles, ","), strings.Join(a.Effective, ","), strings.Join(a.Available, ",")})
		}
		i18n.Printf("Acesso de %s:\n", who)
		table(i18n.T("APLICAÇÃO\tROLES\tEFETIVAS\tDISPONÍVEIS"), rows)
	})
}

func iamAPIKey(ctx context.Context, args []string, flags map[string]bool) error {
	if len(args) < 3 {
		return ui.Usagef("use: romanctl iam api-key create <realm> <nome> | rotate|delete <realm> <id>")
	}
	verb, realm, target := args[0], args[1], args[2]
	base := "/api/realms/" + url.PathEscape(realm) + "/api-keys"
	type issued struct {
		ClientID      string `json:"clientId"`
		Secret        string `json:"secret"`
		TokenEndpoint string `json:"tokenEndpoint"`
	}
	show := func(k issued) {
		i18n.Printf("client_id:      %s\nclient_secret:  %s\ntoken endpoint: %s\n", k.ClientID, k.Secret, k.TokenEndpoint)
		ui.Warnf("O secret só aparece agora; guarde num cofre")
	}
	switch verb {
	case "create":
		body := map[string]any{"name": target, "description": flagVals["description"], "roles": list(flagVals["roles"])}
		var k issued
		if err := iam.Do(ctx, "POST", base, body, &k); err != nil {
			return err
		}
		return output(k, func() { show(k) })
	case "rotate":
		var k issued
		if err := iam.Do(ctx, "POST", base+"/"+url.PathEscape(target)+"/rotate", nil, &k); err != nil {
			return err
		}
		return output(k, func() { show(k) })
	case "delete":
		if err := confirmed(flags, i18n.Sprintf("Excluir a API key %s?", target)); err != nil {
			return err
		}
		if err := iam.Do(ctx, "DELETE", base+"/"+url.PathEscape(target), nil, nil); err != nil {
			return err
		}
		return output(map[string]any{"ok": true}, func() { ui.Okf("API key excluída") })
	}
	return ui.Usagef("use: romanctl iam api-key create <realm> <nome> | rotate|delete <realm> <id>")
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
