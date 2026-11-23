package scaffold

import (
	"strings"

	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/sys"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
)

const (
	setupDone    = "done"
	setupSkipped = "skipped"
	setupFailed  = "failed"
	setupOffline = "offline"
)

// missingTools returns the template's tools that do not run on this machine yet.
func missingTools(keys []string) []string {
	var out []string
	for _, key := range keys {
		t := tools.Find(key)
		if t == nil {
			out = append(out, key)
			continue
		}
		if _, ok := t.Detected(); !ok {
			out = append(out, key)
		}
	}
	return out
}

// ensureTools offers to install what is missing and returns what is still missing.
func ensureTools(keys []string) []string {
	missing := missingTools(keys)
	if len(missing) == 0 {
		return nil
	}
	list := strings.Join(missing, " ")
	if !ui.Interactive() || ui.JSONOutput {
		ui.Warnf("The project needs %s. Install it with: romanctl tools install %s", list, list)
		return missing
	}
	ok, err := ui.Confirm(i18n.Sprintf("The project needs %s. Install it now?", list))
	if err != nil || !ok {
		ui.Warnf("The project needs %s. Install it with: romanctl tools install %s", list, list)
		return missing
	}
	for _, key := range missing {
		if t := tools.Find(key); t != nil {
			tools.InstallOne(t, false)
		}
	}
	return missingTools(keys)
}

// runSetup downloads the new project's dependencies. A failure here does not undo the project.
func runSetup(t *Template, dir string, withs map[string]bool, offline bool, missing []string) string {
	steps := t.setupFor(withs)
	if len(steps) == 0 {
		return ""
	}
	if offline {
		ui.Infof("Offline, dependencies are left for later: cd %s && %s", dir, strings.Join(steps, " && "))
		return setupOffline
	}
	if len(missing) > 0 {
		ui.Infof("Dependencies are left for later: cd %s && %s", dir, strings.Join(steps, " && "))
		return setupSkipped
	}
	for _, command := range steps {
		var err error
		ui.Spin(i18n.Sprintf("Running %s...", command), func(func(string)) { err = sys.Quiet(dir, nil, "sh", "-c", command) })
		if err != nil {
			ui.Warnf("%s failed; the project was created, run it again inside the folder:\n%v", command, err)
			return setupFailed
		}
	}
	return setupDone
}

// Complete fills in, from the template, what an older romanctl.yaml lacks.
func Complete(m *manifest.Project) {
	t := FindByKey(m.Stack, m.Template)
	if t == nil {
		return
	}
	if len(m.Requires) == 0 {
		m.Requires = t.Requires
	}
	if m.Publish.Kind != "" {
		return
	}
	complete := buildManifest(t, m.Name, m.Services)
	m.Publish = complete.Publish
	if m.Env == nil {
		m.Env = complete.Env
	}
}
