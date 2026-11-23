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

// Version is an installable version. ID is what goes to --version and to the installer,
// like the tag "v0.51.0"; Label is what the user reads, like "0.51.0".
type Version struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Note  string `json:"note,omitempty"`
}

const maxVersions = 30

var rePrerelease = regexp.MustCompile(`(?i)(alpha|beta|rc|pre|dev|nightly|snapshot|canary|next|preview|-ent|\+ent)|\d(a|b)\d+$`)

func stable(v string) bool { return !rePrerelease.MatchString(v) }

// compareVersions orders numerically: 1.10.0 comes after 1.9.3.
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

// newestFirst sorts, drops duplicates and trims to what fits in a list.
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

// resolveTag uses the requested tag or, if none, the latest release.
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

// githubVersions reads the releases feed, which does not consume GitHub API rate limit. In a
// repository with several products, like kustomize, it keeps only tags of the same family as
// the latest release.
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
	// The abbreviated document carries only the versions, not each one's README.
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

// pkgVersions asks the package manager which versions the repositories offer.
func pkgVersions(manager string, pkgs []string) ([]Version, error) {
	if len(pkgs) == 0 {
		return nil, i18n.Errorf("no package defined for this distribution")
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
		return nil, i18n.Errorf("%s only installs the repository's current version", manager)
	}
	out := make([]Version, 0, len(ids))
	for _, id := range ids {
		out = append(out, Version{ID: id, Label: id})
	}
	return newestFirst(out), nil
}

func pkgInstallAt(manager string, pkgs []string, version string) error {
	if len(pkgs) == 0 {
		return i18n.Errorf("no package defined for this distribution")
	}
	switch manager {
	case platform.Apt:
		return aptInstall("--allow-downgrades", pkgs[0]+"="+version)
	case platform.Dnf:
		return privileged(nil, "dnf", "install", "-y", pkgs[0]+"-"+version)
	}
	return i18n.Errorf("%s only installs the repository's current version", manager)
}

func nodeVersions() ([]Version, error) {
	var releases []struct {
		Version string `json:"version"`
		LTS     any    `json:"lts"`
	}
	if err := fetchJSON("https://nodejs.org/dist/index.json", &releases); err != nil {
		return nil, err
	}
	// One per major, the newest of each: nobody picks among 40 patches of Node 22.
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
			note = i18n.T("current, not LTS")
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
	// The newest of each line (1.25.x, 1.24.x...).
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

// Versions lists what can be installed on this machine, newest first. The route is the same
// as runInstall's, so the list never offers what the installer cannot install.
func (t *Tool) Versions() ([]Version, error) {
	p := platform.Current
	switch {
	case p.OS == "darwin" && !t.Portable:
		return nil, i18n.Errorf("on macOS, Homebrew only installs the latest version")
	case t.Npm != "":
		return npmVersions(t.Npm)
	case t.PyPI != "":
		return pypiVersions(t.PyPI)
	case t.versions != nil && t.installAt != nil:
		return t.versions()
	case t.Install != nil:
		return nil, i18n.Errorf("the %s installer only brings the latest version", t.Name)
	}
	return pkgVersions(p.Manager, t.pkgsFor(p.Manager))
}

// resolveVersion accepts the label ("0.51.0") or the ID ("v0.51.0") and returns the ID. A version
// outside the short list is still accepted, for whoever knows exactly what they want.
func (t *Tool) resolveVersion(want string) string {
	if list, err := t.Versions(); err == nil {
		for _, v := range list {
			if v.ID == want || v.Label == want || CoreVersion(v.Label) == want {
				return v.ID
			}
		}
		// The tag prefix is what the ID has beyond the label: "v", "kustomize/v" or nothing.
		if len(list) > 0 {
			prefix := strings.TrimSuffix(list[0].ID, list[0].Label)
			if !strings.HasPrefix(want, prefix) {
				return prefix + want
			}
		}
	}
	return want
}

// installVersion installs a chosen version through the same route Versions offers.
func (t *Tool) installVersion(version string) error {
	p := platform.Current
	switch {
	case p.OS == "darwin" && !t.Portable:
		return i18n.Errorf("on macOS, Homebrew only installs the latest version")
	case t.Npm != "":
		return npmInstall(t.Npm+"@"+version, t.RemovePaths)
	case t.PyPI != "" && t.venvDest != "":
		return installVenv(t.venvDest, t.PyPI+"=="+version, t.venvBins)
	case t.installAt != nil:
		return t.installAt(version)
	case t.Install != nil:
		return i18n.Errorf("the %s installer only brings the latest version", t.Name)
	}
	return pkgInstallAt(p.Manager, t.pkgsFor(p.Manager), version)
}
