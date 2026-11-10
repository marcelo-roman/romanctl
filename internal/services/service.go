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

	Unsupported string          // por que não roda nesta máquina
	LogCmd      func() []string // comando que segue o log, quando não há arquivo
	LogTail     func() []string // o mesmo, só com as últimas linhas e sem seguir
	Engine      bool            // é o motor de containers: fica fora de "all"
	before      func() error    // roda antes do spinner, como a senha do sudo
	Warning     string          // o que mais para junto; parar pede confirmação
	Conflict    func() string   // outro programa disputando a porta do serviço já ligado
}

// All lista os serviços: no proot, Postgres, Mongo e Redis rodam como processo local;
// nas demais máquinas, tudo vai em container.
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
		return i18n.Errorf("%s não está instalado — rode: romanctl install %s", s.Name, s.Key)
	}
	state.MarkService(s.Key, true) // mesmo se falhar: o restore tenta de novo
	if s.Running() {
		ui.Okf("%s já está rodando → %s", s.Name, s.URL())
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
	ui.Spin(i18n.Sprintf("Iniciando %s...", s.Name), func(func(string)) {
		if startErr = s.start(); startErr != nil {
			startErr = i18n.Errorf("%s: %w (veja: romanctl logs %s)", s.Name, startErr, s.Key)
			return
		}
		if !sys.WaitUntil(30*time.Second, s.Running) {
			startErr = i18n.Errorf("%s não subiu a tempo (veja: romanctl logs %s)", s.Name, s.Key)
		}
	})
	if startErr != nil {
		return startErr
	}
	ui.Okf("%s rodando → %s", s.Name, s.URL())
	return nil
}

func (s *Service) Stop() error {
	state.MarkService(s.Key, false)
	if !s.Running() {
		ui.Infof("%s não estava rodando", s.Name)
		return nil
	}
	if s.before != nil {
		if err := s.before(); err != nil {
			return err
		}
	}
	var stopErr error
	ui.Spin(i18n.Sprintf("Parando %s...", s.Name), func(func(string)) {
		if stopErr = s.stop(); stopErr != nil {
			stopErr = i18n.Errorf("%s: %w", s.Name, stopErr)
			return
		}
		if !sys.WaitUntil(20*time.Second, func() bool { return !s.Running() }) {
			stopErr = i18n.Errorf("%s não parou a tempo", s.Name)
		}
	})
	if stopErr != nil {
		return stopErr
	}
	ui.Okf("%s parado", s.Name)
	return nil
}

func (s *Service) Install() error {
	if s.Unsupported == "" && s.Installed() {
		ui.Okf("%s já está instalado", s.Name)
		return nil
	}
	ui.Infof("Instalando %s (precisa de internet)...", s.Name)
	if err := s.install(); err != nil {
		return i18n.Errorf("instalação de %s falhou: %w", s.Name, err)
	}
	ui.Okf("%s instalado", s.Name)
	return nil
}

func aptInstall(pkgs ...string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	return sys.Run("", []string{"DEBIAN_FRONTEND=noninteractive"}, "apt-get", args...)
}
