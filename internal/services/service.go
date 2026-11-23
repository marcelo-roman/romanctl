package services

import (
	"romanctl/internal/i18n"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
	"time"
)

type Service struct {
	Key       string
	Name      string
	Aliases   []string
	URL       func() string
	LogFile   string
	Installed func() bool
	Running   func() bool
	start     func() error
	stop      func() error
	install   func() error

	Unsupported string          // why it does not run on this machine
	LogCmd      func() []string // command that follows the log, when there is no file
	LogTail     func() []string // the same, only the last lines and without following
	Engine      bool            // it is the container engine: left out of "all"
	before      func() error    // runs before the spinner, like the sudo password
	Warning     string          // what else stops with it; stopping asks for confirmation
	Conflict    func() string   // another program holding the port of the running service
}

// All lists the services: under proot, Postgres, Mongo and Redis run as local processes;
// on every other machine, everything runs in a container.
var All = append([]*Service{dockerEngine(), codeServer(), postgres(), mongo(), redis()}, containerOnly()...)

func FindService(name string) *Service {
	name = strings.ToLower(name)
	for _, s := range All {
		if s.Key == name {
			return s
		}
		for _, a := range s.Aliases {
			if a == name {
				return s
			}
		}
	}
	return nil
}

func (s *Service) Start() error {
	if s.Unsupported != "" {
		return i18n.Errorf("%s: %s", s.Name, s.Unsupported)
	}
	if !s.Installed() {
		return i18n.Errorf("%s is not installed — run: romanctl install %s", s.Name, s.Key)
	}
	state.MarkService(s.Key, true) // even if it fails: restore tries again
	if s.Running() {
		ui.Okf("%s is already running → %s", s.Name, s.URL())
		if s.Conflict != nil {
			if warning := s.Conflict(); warning != "" {
				ui.Warnf("%s", warning)
			}
		}
		return nil
	}
	if s.before != nil {
		if err := s.before(); err != nil {
			return err
		}
	}
	var startErr error
	ui.Spin(i18n.Sprintf("Starting %s...", s.Name), func(func(string)) {
		if startErr = s.start(); startErr != nil {
			startErr = i18n.Errorf("%s: %w (see: romanctl logs %s)", s.Name, startErr, s.Key)
			return
		}
		if !sys.WaitUntil(30*time.Second, s.Running) {
			startErr = i18n.Errorf("%s did not start in time (see: romanctl logs %s)", s.Name, s.Key)
		}
	})
	if startErr != nil {
		return startErr
	}
	ui.Okf("%s running → %s", s.Name, s.URL())
	return nil
}

func (s *Service) Stop() error {
	state.MarkService(s.Key, false)
	if !s.Running() {
		ui.Infof("%s was not running", s.Name)
		return nil
	}
	if s.before != nil {
		if err := s.before(); err != nil {
			return err
		}
	}
	var stopErr error
	ui.Spin(i18n.Sprintf("Stopping %s...", s.Name), func(func(string)) {
		if stopErr = s.stop(); stopErr != nil {
			stopErr = i18n.Errorf("%s: %w", s.Name, stopErr)
			return
		}
		if !sys.WaitUntil(20*time.Second, func() bool { return !s.Running() }) {
			stopErr = i18n.Errorf("%s did not stop in time", s.Name)
		}
	})
	if stopErr != nil {
		return stopErr
	}
	ui.Okf("%s stopped", s.Name)
	return nil
}

func (s *Service) Install() error {
	if s.Unsupported == "" && s.Installed() {
		ui.Okf("%s is already installed", s.Name)
		return nil
	}
	ui.Infof("Installing %s (needs internet)...", s.Name)
	if err := s.install(); err != nil {
		return i18n.Errorf("installation of %s failed: %w", s.Name, err)
	}
	ui.Okf("%s installed", s.Name)
	return nil
}

func aptInstall(pkgs ...string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	return sys.Run("", []string{"DEBIAN_FRONTEND=noninteractive"}, "apt-get", args...)
}
