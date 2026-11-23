package tools

import (
	"encoding/json"
	"regexp"
	"strings"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
)

// Upstream is what installing would bring right now: the version, where it comes from and the project's README.
type Upstream struct {
	Latest string `json:"latest,omitempty"`
	Source string `json:"source"`
	URL    string `json:"url,omitempty"`
	Readme string `json:"-"`
}

// from points to the GitHub project whose releases are what the installer downloads.
func (t *Tool) from(repo string) *Tool {
	t.Repo = repo
	t.latest = func() (string, error) { return githubLatestVersion(repo) }
	return t
}

type source struct {
	repo     string
	release  bool
	latest   func() (string, error)
	pypi     string
	install  func(version string) error
	versions func() ([]Version, error)
}

func github(repo string, install func(string) error) source {
	return source{repo: repo, release: true, latest: func() (string, error) { return githubLatestVersion(repo) },
		install: install, versions: func() ([]Version, error) { return githubVersions(repo) }}
}

// sources complements the custom installers with their origin. latest is the same query the
// installer makes, so the version shown is the one that will arrive.
var sources = map[string]source{
	"kubectl":     {repo: "kubernetes/kubectl", latest: kubectlLatest, install: installKubectlAt, versions: kubectlVersions},
	"helm":        {repo: "helm/helm", latest: func() (string, error) { return githubLatestVersion("helm/helm") }, install: installHelmAt, versions: func() ([]Version, error) { return githubVersions("helm/helm") }},
	"node":        {repo: "nodejs/node", latest: nodeLTS, install: installNodeAt, versions: nodeVersions},
	"go":          {repo: "golang/go", latest: goLatest, install: installGoAt, versions: goVersions},
	"zig":         {repo: "ziglang/zig", latest: zigLatest, install: installZigAt, versions: zigVersions},
	"rust":        {repo: "rust-lang/rust"},
	"dart":        {repo: "dart-lang/sdk"},
	"swift":       {repo: "swiftlang/swift"},
	"aws":         {repo: "aws/aws-cli"},
	"az":          {repo: "Azure/azure-cli", pypi: "azure-cli"},
	"docker-cli":  {repo: "docker/cli"},
	"neovim":      github("neovim/neovim", installNeovimAt),
	"ollama":      github("ollama/ollama", installOllamaAt),
	"kotlin":      github("JetBrains/kotlin", installKotlinAt),
	"lima":        github("lima-vm/lima", installLimaAt),
	"code-server": github("coder/code-server", installCodeServerAt),
}

func init() {
	for _, t := range All {
		s, ok := sources[t.Key]
		if !ok {
			continue
		}
		t.Repo, t.Release, t.latest, t.PyPI = s.repo, s.release, s.latest, s.pypi
		if s.install != nil {
			t.installAt, t.versions = s.install, s.versions
		}
	}
}

// Upstream queries the source runInstall would use on this machine. Network failures leave the
// fields empty: the information helps, but must not block the installation.
func (t *Tool) Upstream(withReadme bool) Upstream {
	p := platform.Current
	var u Upstream
	switch {
	case p.OS == "darwin" && !t.Portable:
		u = brewUpstream(t.Brew, t.Cask)
	case t.Npm != "":
		u = npmUpstream(t.Npm)
	case t.PyPI != "":
		u = pypiUpstream(t.PyPI, withReadme)
	case t.Install != nil:
		u.Source = i18n.T("official installer")
		if t.Release {
			u.Source = "GitHub releases"
		}
		if t.latest != nil {
			u.Latest, _ = t.latest()
		}
	default:
		u = pkgUpstream(p.Manager, t.pkgsFor(p.Manager))
	}
	repo := u.Repo()
	if repo == "" {
		repo = t.Repo
	}
	if u.URL == "" && repo != "" {
		u.URL = "https://github.com/" + repo
	}
	if withReadme && u.Readme == "" && repo != "" {
		u.Readme = githubReadme(repo)
	}
	u.Latest = strings.TrimPrefix(strings.TrimSpace(u.Latest), "v")
	return u
}

var reGitHub = regexp.MustCompile(`github\.com[/:]([\w.-]+/[\w.-]+?)(?:\.git)?(?:[/#?]|$)`)

// Repo returns the GitHub project, when the page points there.
func (u Upstream) Repo() string {
	if m := reGitHub.FindStringSubmatch(u.URL); m != nil {
		return m[1]
	}
	return ""
}

var reVersion = regexp.MustCompile(`\d+(?:\.\d+)+`)

// CoreVersion strips what is not numeric, to compare "helm v3.16.2+g1a2b" with "3.16.2".
func CoreVersion(s string) string { return reVersion.FindString(s) }

// UpToDate says whether the installed version is the one installing would bring.
func UpToDate(installed, latest string) bool {
	a, b := CoreVersion(installed), CoreVersion(latest)
	return a != "" && a == b
}

func githubLatestVersion(repo string) (string, error) {
	tag, err := githubLatestTag(repo)
	if err != nil {
		return "", err
	}
	return tag[strings.LastIndex(tag, "/")+1:], nil
}

func githubReadme(repo string) string {
	for _, name := range []string{"README.md", "readme.md", "Readme.md", "README.rst", "README"} {
		if text, err := fetchText("https://raw.githubusercontent.com/" + repo + "/HEAD/" + name); err == nil && text != "" {
			return text
		}
	}
	return ""
}

func brewUpstream(names []string, cask bool) Upstream {
	u := Upstream{Source: "Homebrew"}
	if len(names) == 0 {
		return u
	}
	args := []string{"info", "--json=v2"}
	if cask {
		args = append(args, "--cask")
	}
	out, err := sys.Output("brew", append(args, names[0])...)
	if err != nil {
		return u
	}
	var info struct {
		Formulae []struct {
			Homepage string `json:"homepage"`
			Versions struct {
				Stable string `json:"stable"`
			} `json:"versions"`
		} `json:"formulae"`
		Casks []struct {
			Homepage string `json:"homepage"`
			Version  string `json:"version"`
		} `json:"casks"`
	}
	if json.Unmarshal([]byte(out), &info) != nil {
		return u
	}
	switch {
	case len(info.Formulae) > 0:
		u.Latest, u.URL = info.Formulae[0].Versions.Stable, info.Formulae[0].Homepage
	case len(info.Casks) > 0:
		u.Latest, u.URL = info.Casks[0].Version, info.Casks[0].Homepage
	}
	return u
}

func npmUpstream(pkg string) Upstream {
	u := Upstream{Source: "npm " + pkg, URL: "https://www.npmjs.com/package/" + pkg}
	var info struct {
		Version    string `json:"version"`
		Homepage   string `json:"homepage"`
		Repository any    `json:"repository"`
	}
	if fetchJSON("https://registry.npmjs.org/"+pkg+"/latest", &info) != nil {
		return u
	}
	u.Latest = info.Version
	repo := ""
	switch r := info.Repository.(type) {
	case string:
		repo = r
	case map[string]any:
		repo, _ = r["url"].(string)
	}
	for _, candidate := range []string{repo, info.Homepage} {
		if reGitHub.MatchString(candidate) {
			u.URL = candidate
			break
		}
	}
	return u
}

func pypiUpstream(pkg string, withReadme bool) Upstream {
	u := Upstream{Source: "PyPI " + pkg, URL: "https://pypi.org/project/" + pkg}
	var info struct {
		Info struct {
			Version     string            `json:"version"`
			Description string            `json:"description"`
			ProjectURLs map[string]string `json:"project_urls"`
		} `json:"info"`
	}
	if fetchJSON("https://pypi.org/pypi/"+pkg+"/json", &info) != nil {
		return u
	}
	u.Latest = info.Info.Version
	if withReadme {
		u.Readme = info.Info.Description
	}
	for _, link := range info.Info.ProjectURLs {
		if reGitHub.MatchString(link) {
			u.URL = link
			break
		}
	}
	return u
}

var (
	reAptCandidate = regexp.MustCompile(`(?m)^\s*Candidate:\s*(\S+)`)
	reFieldVersion = regexp.MustCompile(`(?m)^Version\s*:\s*(\S+)`)
	reApkVersion   = regexp.MustCompile(`(?m)^\s+(\S+):\s*$`)
)

// pkgUpstream asks the package manager for the candidate version, without installing anything.
func pkgUpstream(manager string, pkgs []string) Upstream {
	u := Upstream{Source: manager}
	if len(pkgs) == 0 {
		return u
	}
	pkg := pkgs[0]
	var out string
	var re *regexp.Regexp
	switch manager {
	case platform.Apt:
		out, _ = sys.Output("apt-cache", "policy", pkg)
		re = reAptCandidate
	case platform.Dnf:
		out, _ = sys.Output("dnf", "repoquery", "--latest-limit=1", "--qf", "%{version}\n", pkg)
		u.Latest = strings.TrimSpace(strings.SplitN(out, "\n", 2)[0])
		return u
	case platform.Pacman:
		out, _ = sys.Output("pacman", "-Si", pkg)
		re = reFieldVersion
	case platform.Zypper:
		out, _ = sys.Output("zypper", "--no-refresh", "info", pkg)
		re = reFieldVersion
	case platform.Apk:
		out, _ = sys.Output("apk", "policy", pkg)
		re = reApkVersion
	default:
		return u
	}
	if m := re.FindStringSubmatch(out); m != nil && m[1] != "(none)" {
		u.Latest = m[1]
	}
	return u
}
