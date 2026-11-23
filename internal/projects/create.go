package projects

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/services"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
)

type ProjectType struct {
	Key    string
	Label  string
	Create func(name, path string) error
}

var ProjectTypes = []ProjectType{
	{"node", i18n.T("Node.js (basic)"), createNode(false)},
	{"express", "API Node.js + Express", createNode(true)},
	{"react", "React + Vite (JavaScript)", createVite("react")},
	{"react-ts", "React + Vite (TypeScript)", createVite("react-ts")},
	{"next", "Next.js (TypeScript, App Router, Tailwind)", createNext},
	{"dotnet-webapi", ".NET Web API", createDotnet("webapi")},
	{"dotnet-console", ".NET Console", createDotnet("console")},
	{"dotnet-mvc", ".NET MVC", createDotnet("mvc")},
	{"dotnet-blazor", ".NET Blazor", createDotnet("blazor")},
	{"go", i18n.T("Go (module + main.go)"), createGo},
}

func FindType(k string) *ProjectType {
	for i := range ProjectTypes {
		if ProjectTypes[i].Key == strings.ToLower(k) {
			return &ProjectTypes[i]
		}
	}
	return nil
}

var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

func ValidateName(n string) error {
	if !nameRe.MatchString(n) {
		return i18n.Errorf("use lowercase letters, numbers, '-', '_' or '.' (e.g. my-app)")
	}
	if sys.Exists(filepath.Join(config.ProjDir, n)) {
		return i18n.Errorf("a project named %q already exists", n)
	}
	return nil
}

func NewProject(typ, name string, git, forceOffline bool) error {
	t := FindType(typ)
	if t == nil {
		return i18n.Errorf("unknown type %q (types: %s)", typ, TypeKeys())
	}
	if err := ValidateName(name); err != nil {
		return err
	}
	os.MkdirAll(config.ProjDir, 0o755)
	path := filepath.Join(config.ProjDir, name)
	ui.Infof("Creating %s in %s ...", t.Label, path)
	offlineMode = forceOffline || !sys.IsOnline()
	if offlineMode {
		ui.Warnf("No internet: using local templates and cache")
	}
	var err error
	if offlineMode && isTemplated(t.Key) {
		err = createFromTemplate(t.Key, name, path)
	} else {
		err = t.Create(name, path)
		if err != nil && !offlineMode && isTemplated(t.Key) && HasTemplate(t.Key) {
			ui.Warnf("The generator failed (%v). Trying the local template...", err)
			os.RemoveAll(path)
			err = createFromTemplate(t.Key, name, path)
		}
	}
	if err != nil {
		os.RemoveAll(path) // never leave a half-created project
		return i18n.Errorf("failed to create %s: %w", name, err)
	}
	if git && sys.Have("git") && !sys.Exists(filepath.Join(path, ".git")) {
		sys.Output("git", "-C", path, "init", "-q")
	}
	ui.Okf("Project %s created!", ui.SBoldTxt.Render(name))
	i18n.Printf("   Folder:  %s\n   Run:     romanctl run %s\n   Open:    %s/?folder=%s\n",
		path, name, services.CodeServerURL(), path)
	return nil
}

func TypeKeys() string {
	var ks []string
	for _, t := range ProjectTypes {
		ks = append(ks, t.Key)
	}
	return strings.Join(ks, ", ")
}

func createNode(express bool) func(string, string) error {
	return func(name, path string) error {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
		if out, err := sys.OutputIn(path, "npm", "init", "-y"); err != nil {
			return i18n.Errorf("npm init: %v: %s", err, out)
		}
		index := "console.log('Hello from " + name + "!');\n"
		if express {
			if err := npmInstall(path, "express"); err != nil {
				return err
			}
			index = `const express = require('express');

const app = express();
app.use(express.json());

app.get('/', (req, res) => {
  res.json({ message: 'Hello from ` + name + `!' });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log('API rodando em http://localhost:' + PORT));
`
		}
		os.WriteFile(filepath.Join(path, "index.js"), []byte(index), 0o644)
		os.WriteFile(filepath.Join(path, ".gitignore"), []byte("node_modules/\n.env\n"), 0o644)
		return setScripts(filepath.Join(path, "package.json"), map[string]string{
			"dev": "node --watch index.js", "start": "node index.js"})
	}
}

func setScripts(pkgPath string, scripts map[string]string) error {
	b, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}
	var pkg map[string]any
	if err := json.Unmarshal(b, &pkg); err != nil {
		return err
	}
	sc, _ := pkg["scripts"].(map[string]any)
	if sc == nil {
		sc = map[string]any{}
	}
	delete(sc, "test")
	for k, v := range scripts {
		sc[k] = v
	}
	pkg["scripts"] = sc
	out, _ := json.MarshalIndent(pkg, "", "  ")
	return os.WriteFile(pkgPath, append(out, '\n'), 0o644)
}

func createVite(tpl string) func(string, string) error {
	return func(name, path string) error {
		if err := sys.Run(filepath.Dir(path), nil, "npm", "create", "vite@latest", filepath.Base(path), "--", "--template", tpl, "--no-interactive"); err != nil {
			return err
		}
		return sys.Run(path, nil, "npm", "install")
	}
}

// The default layout downloads Google fonts at build time, which breaks "next build"
// offline. The geist package ships the same font as local files.
const nextLayout = `import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import "./globals.css";

export const metadata: Metadata = {
  title: "Create Next App",
  description: "Generated by create next app",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html
      lang="pt-BR"
      className={` + "`${GeistSans.variable} ${GeistMono.variable} h-full antialiased`" + `}
    >
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
`

func createNext(name, path string) error {
	if err := sys.Run(filepath.Dir(path), []string{"NEXT_TELEMETRY_DISABLED=1"}, "npx", "--yes", "create-next-app@latest", filepath.Base(path),
		"--ts", "--app", "--tailwind", "--eslint", "--src-dir", "--use-npm", "--disable-git", "--import-alias", "@/*", "--yes"); err != nil {
		return err
	}
	if err := npmInstall(path, "geist"); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "src", "app", "layout.tsx"), []byte(nextLayout), 0o644)
}

func createDotnet(tpl string) func(string, string) error {
	return func(name, path string) error {
		if err := sys.Run(filepath.Dir(path), []string{"DOTNET_NOLOGO=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1"},
			"dotnet", "new", tpl, "-o", path, "-n", DotnetName(name)); err != nil {
			return err
		}
		if !sys.Exists(filepath.Join(path, ".gitignore")) {
			sys.Output("dotnet", "new", "gitignore", "-o", path)
		}
		return nil
	}
}

// .NET names do not accept '-' in the namespace; my-app becomes MyApp.
func DotnetName(n string) string {
	parts := strings.FieldsFunc(n, func(r rune) bool { return r == '-' || r == '_' || r == '.' })
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func createGo(name, path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	if out, err := sys.OutputIn(path, "go", "mod", "init", name); err != nil {
		return i18n.Errorf("%v: %s", err, out)
	}
	main := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello from " + name + "!\")\n}\n"
	return os.WriteFile(filepath.Join(path, "main.go"), []byte(main), 0o644)
}
