package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

var (
	spinMu     sync.Mutex
	spinActive bool
	spinText   string
)

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Só anima em um terminal: um pipe ou --json nunca recebem esses caracteres.
func spinnerEnabled() bool {
	return !silent && !JSONOutput && term.IsTerminal(int(os.Stderr.Fd()))
}

// Spin mostra uma animação enquanto fn roda. O aviso só aparece se a espera passar de um
// instante, para operações rápidas não piscarem. fn pode trocar o texto com update.
func Spin(message string, fn func(update func(string))) {
	spinMu.Lock()
	nested := spinActive
	spinMu.Unlock()
	if !spinnerEnabled() || nested {
		fn(func(string) {})
		return
	}
	spinMu.Lock()
	spinActive, spinText = true, message
	spinMu.Unlock()

	done, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		time.Sleep(120 * time.Millisecond)
		tick := time.NewTicker(90 * time.Millisecond)
		defer tick.Stop()
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			case <-tick.C:
				spinMu.Lock()
				fmt.Fprintf(os.Stderr, "\r\033[K%s %s", SBlue.Render(spinFrames[i%len(spinFrames)]), spinText)
				spinMu.Unlock()
			}
		}
	}()
	defer func() {
		close(done)
		<-stopped
		spinMu.Lock()
		spinActive = false
		fmt.Fprint(os.Stderr, "\r\033[K")
		spinMu.Unlock()
	}()
	fn(func(text string) {
		spinMu.Lock()
		spinText = text
		spinMu.Unlock()
	})
}

// clearSpinnerLine abre espaço para uma mensagem aparecer acima da animação.
func clearSpinnerLine() {
	if spinActive {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}
}
