package tools

import (
	"cmp"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
)

// Version é uma versão que dá para instalar. ID é o que vai para --version e para o
// instalador, como a tag "v0.51.0"; Label é o que a pessoa lê, como "0.51.0".
type Version struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Note  string `json:"note,omitempty"`
}

const maxVersions = 30

var rePrerelease = regexp.MustCompile(`(?i)(alpha|beta|rc|pre|dev|nightly|snapshot|canary|next|preview|-ent|\+ent)|\d(a|b)\d+$`)

func stable(v string) bool { return !rePrerelease.MatchString(v) }

// compareVersions ordena pelos números: 1.10.0 vem depois de 1.9.3.
func compareVersions(a, b string) int {
	x, y := reVersion.FindString(a), reVersion.FindString(b)
	xs, ys := strings.Split(x, "."), strings.Split(y, ".")
	for i := 0; i < max(len(xs), len(ys)); i++ {
		var p, q int
		if i < len(xs) {
			p, _ = strconv.Atoi(xs[i])
		}
		if i < len(ys) {
			q, _ = strconv.Atoi(ys[i])
		}
		if p != q {
			return cmp.Compare(p, q)
		}
	}
	return 0
}

// newestFirst ordena, tira as repetidas e corta na quantidade que cabe numa lista.
func newestFirst(list []Version) []Version {
	slices.SortStableFunc(list, func(a, b Version) int { return compareVersions(b.Label, a.Label) })
	list = slices.CompactFunc(list, func(a, b Version) bool { return a.ID == b.ID })
	return list[:min(len(list), maxVersions)]
}

func plain(ids []string) []Version {
	out := make([]Version, 0, len(ids))
	for _, id := range ids {
		if stable(id) && CoreVersion(id) != "" {
			out = append(out, Version{ID: id, Label: strings.TrimPrefix(id, "v")})
		}
	}
	return newestFirst(out)
}

// resolveTag usa a tag pedida ou, sem pedido, a última release.
func resolveTag(repo, want string) (string, error) {
	if want != "" {
		return want, nil
	}
	return githubLatestTag(repo)
}

func orLatest(want string, latest func() (string, error)) (string, error) {
	if want != "" {
		return want, nil
	}
	return latest()
}

var reReleaseTag = regexp.MustCompile(`/releases/tag/([^"]+)"`)

// githubVersions lê o feed de releases, que não gasta o limite da API do GitHub. Num
// repositório com vários produtos, como o kustomize, fica só com as tags da mesma família
// da última release.
func githubVersions(repo string) ([]Version, error) {
	feed, err := fetchText("https://github.com/" + repo + "/releases.atom")
	if err != nil {
		return nil, err
	}
	latest, _ := githubLatestTag(repo)
	family := ""
	if i := strings.LastIndex(latest, "/"); i >= 0 {
		family = latest[:i+1]
	}
	var out []Version
	for _, m := range reReleaseTag.FindAllStringSubmatch(feed, -1) {
		tag, err := url.PathUnescape(m[1])
		if err != nil || !strings.HasPrefix(tag, family) || !stable(tag) || CoreVersion(tag) == "" {
			continue
		}
		out = append(out, Version{ID: tag, Label: strings.TrimPrefix(strings.TrimPrefix(tag, family), "v")})
	}
	return newestFirst(out), nil
}

func hashicorpVersions(key string) ([]Version, error) {
	var index struct {
		Versions map[string]any `json:"versions"`
	}
	if err := fetchJSON("https://releases.hashicorp.com/"+key+"/index.json", &index); err != nil {
		return nil, err
	}
	var ids []string
	for v := range index.Versions {
		if !strings.ContainsAny(v, "-+") {
			ids = append(ids, v)
		}
	}
	return plain(ids), nil
}

func npmVersions(pkg string) ([]Version, error) {
	var doc struct {
		Versions map[string]any `json:"versions"`
	}
	// O documento resumido traz só as versões, e não o README de cada uma.
	if err := fetchJSONWith("https://registry.npmjs.org/"+pkg, &doc, "Accept", "application/vnd.npm.install-v1+json"); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(doc.Versions))
	for v := range doc.Versions {
		ids = append(ids, v)
	}
	return plain(ids), nil
}

func pypiVersions(pkg string) ([]Version, error) {
	var doc struct {
		Releases map[string][]struct {
			Yanked bool `json:"yanked"`
		} `json:"releases"`
	}
	if err := fetchJSON("https://pypi.org/pypi/"+pkg+"/json", &doc); err != nil {
		return nil, err
	}
	var ids []string
	for v, files := range doc.Releases {
		if len(files) > 0 && !files[0].Yanked {
			ids = append(ids, v)
		}
	}
	return plain(ids), nil
}

// pkgVersions pergunta ao gerenciador de pacotes quais versões os repositórios oferecem.
func pkgVersions(manager string, pkgs []string) ([]Version, error) {
	if len(pkgs) == 0 {
		return nil, i18n.Errorf("nenhum pacote definido para esta distribuição")
	}
	var ids []string
	switch manager {
	case platform.Apt:
		out, err := sys.Output("apt-cache", "madison", pkgs[0])
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(out, "\n") {
			if f := strings.Split(line, "|"); len(f) >= 2 {
				ids = append(ids, strings.TrimSpace(f[1]))
			}
		}
	case platform.Dnf:
		out, err := sys.Output("dnf", "list", "--showduplicates", "--available", "-q", pkgs[0])
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(out, "\n") {
			if f := strings.Fields(line); len(f) >= 2 && strings.HasPrefix(f[0], pkgs[0]+".") {
				ids = append(ids, f[1])
			}
		}
	default:
		return nil, i18n.Errorf("o %s instala só a versão atual do repositório", manager)
	}
	out := make([]Version, 0, len(ids))
	for _, id := range ids {
		out = append(out, Version{ID: id, Label: id})
	}
	return newestFirst(out), nil
}

func pkgInstallAt(manager string, pkgs []string, version string) error {
	if len(pkgs) == 0 {
		return i18n.Errorf("nenhum pacote definido para esta distribuição")
	}
	switch manager {
	case platform.Apt:
		return aptInstall("--allow-downgrades", pkgs[0]+"="+version)
	case platform.Dnf:
		return privileged(nil, "dnf", "install", "-y", pkgs[0]+"-"+version)
	}
	return i18n.Errorf("o %s instala só a versão atual do repositório", manager)
}

func nodeVersions() ([]Version, error) {
	var releases []struct {
		Version string `json:"version"`
		LTS     any    `json:"lts"`
	}
	if err := fetchJSON("https://nodejs.org/dist/index.json", &releases); err != nil {
		return nil, err
	}
	// Uma por major, a mais nova de cada: ninguém escolhe entre 40 patches do Node 22.
	seen := map[string]bool{}
	var out []Version
	for i, r := range releases {
		major := strings.SplitN(strings.TrimPrefix(r.Version, "v"), ".", 2)[0]
		if seen[major] || !stable(r.Version) {
			continue
		}
		note := ""
		switch lts := r.LTS.(type) {
		case string:
			note = "LTS " + lts
		default:
			if len(out) > 0 || i > 3 {
				continue
			}
			note = i18n.T("atual, sem LTS")
		}
		seen[major] = true
		out = append(out, Version{ID: r.Version, Label: strings.TrimPrefix(r.Version, "v"), Note: note})
	}
	return out[:min(len(out), maxVersions)], nil
}

func goVersions() ([]Version, error) {
	releases, err := goAllReleases()
	if err != nil {
		return nil, err
	}
	// A mais nova de cada linha (1.25.x, 1.24.x...).
	seen := map[string]bool{}
	var out []Version
	for _, r := range releases {
		v := strings.TrimPrefix(r.Version, "go")
		if !stable(v) {
			continue
		}
		parts := strings.SplitN(v, ".", 3)
		line := strings.Join(parts[:min(2, len(parts))], ".")
		if seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, Version{ID: v, Label: v})
	}
	return newestFirst(out), nil
}

func zigVersions() ([]Version, error) {
	index, _, err := zigIndex()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(index))
	for v := range index {
		if v != "master" {
			ids = append(ids, v)
		}
	}
	return plain(ids), nil
}

func kubectlVersions() ([]Version, error) { return githubVersions("kubernetes/kubernetes") }

// Versions lista o que dá para instalar nesta máquina, da mais nova para a mais antiga. A
// rota é a mesma de runInstall, então a lista nunca oferece o que o instalador não sabe pôr.
func (t *Tool) Versions() ([]Version, error) {
	p := platform.Current
	switch {
	case p.OS == "darwin" && !t.Portable:
		return nil, i18n.Errorf("no macOS, o Homebrew instala só a última versão")
	case t.Npm != "":
		return npmVersions(t.Npm)
	case t.PyPI != "":
		return pypiVersions(t.PyPI)
	case t.versions != nil && t.installAt != nil:
		return t.versions()
	case t.Install != nil:
		return nil, i18n.Errorf("o instalador de %s só traz a última versão", t.Name)
	}
	return pkgVersions(p.Manager, t.pkgsFor(p.Manager))
}

// resolveVersion aceita o rótulo ("0.51.0") ou o ID ("v0.51.0") e devolve o ID. Uma versão
// fora da lista curta ainda é aceita, para quem sabe exatamente o que quer.
func (t *Tool) resolveVersion(want string) string {
	if list, err := t.Versions(); err == nil {
		for _, v := range list {
			if v.ID == want || v.Label == want || CoreVersion(v.Label) == want {
				return v.ID
			}
		}
		// O prefixo da tag é o que o ID tem a mais que o rótulo: "v", "kustomize/v" ou nada.
		if len(list) > 0 {
			prefix := strings.TrimSuffix(list[0].ID, list[0].Label)
			if !strings.HasPrefix(want, prefix) {
				return prefix + want
			}
		}
	}
	return want
}

// installVersion instala uma versão escolhida pela mesma rota que Versions oferece.
func (t *Tool) installVersion(version string) error {
	p := platform.Current
	switch {
	case p.OS == "darwin" && !t.Portable:
		return i18n.Errorf("no macOS, o Homebrew instala só a última versão")
	case t.Npm != "":
		return npmInstall(t.Npm+"@"+version, t.RemovePaths)
	case t.PyPI != "" && t.venvDest != "":
		return installVenv(t.venvDest, t.PyPI+"=="+version, t.venvBins)
	case t.installAt != nil:
		return t.installAt(version)
	case t.Install != nil:
		return i18n.Errorf("o instalador de %s só traz a última versão", t.Name)
	}
	return pkgInstallAt(p.Manager, t.pkgsFor(p.Manager), version)
}
