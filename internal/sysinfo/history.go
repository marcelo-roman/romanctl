package sysinfo

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"romanctl/internal/i18n"
	"time"

	"romanctl/internal/config"
)

const historyLimit = 3000

// Reading é uma linha do histórico de bateria.
type Reading struct {
	At      time.Time `json:"at"`
	Percent float64   `json:"percent"`
	Status  string    `json:"status"`
	Charge  float64   `json:"chargeAh,omitempty"`
	Power   float64   `json:"powerW,omitempty"`
}

func historyPath() string { return filepath.Join(config.DataDir, "bateria.jsonl") }

// RecordAndRate grava a leitura e completa a bateria com a queda real medida.
func RecordAndRate(b *Battery, now time.Time) {
	history := loadHistory()
	history = append(history, Reading{now, b.Percent, b.Status, b.ChargeAh, b.Power})
	if len(history) > historyLimit {
		history = history[len(history)-historyLimit:]
	}
	saveHistory(history)

	rate, window := dischargeRate(history)
	if rate <= 0 {
		return
	}
	b.RatePerHour, b.RateWindowMins = rate, window
	if b.Status == "discharging" && b.HoursLeft == 0 {
		b.HoursLeft, b.HoursBasis = b.Percent/rate, i18n.T("queda medida pelo histórico")
	}
}

func loadHistory() []Reading {
	f, err := os.Open(historyPath())
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []Reading
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r Reading
		if json.Unmarshal(sc.Bytes(), &r) == nil {
			out = append(out, r)
		}
	}
	return out
}

func saveHistory(history []Reading) {
	if err := os.MkdirAll(filepath.Dir(historyPath()), 0o755); err != nil {
		return
	}
	tmp := historyPath() + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return
	}
	enc := json.NewEncoder(f)
	for _, r := range history {
		enc.Encode(r)
	}
	if f.Close() == nil {
		os.Rename(tmp, historyPath())
	}
}

// dischargeRate mede a queda em %/h no trecho mais recente de descarga contínua.
// Exige ao menos 15 minutos e 1%, porque abaixo disso o resultado é só ruído.
func dischargeRate(history []Reading) (perHour, minutes float64) {
	if len(history) < 2 {
		return 0, 0
	}
	last := history[len(history)-1]
	if last.Status != "discharging" {
		return 0, 0
	}
	first := last
	for i := len(history) - 2; i >= 0; i-- {
		r := history[i]
		if r.Status != "discharging" || first.At.Sub(r.At) > 2*time.Hour || r.Percent < first.Percent {
			break
		}
		first = r
	}
	elapsed := last.At.Sub(first.At)
	drop := first.Percent - last.Percent
	if elapsed < 15*time.Minute || drop < 1 {
		return 0, 0
	}
	return drop / elapsed.Hours(), elapsed.Minutes()
}
