// Package sysinfo lê CPU, memória, disco, processos e bateria da máquina.
package sysinfo

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"romanctl/internal/i18n"
	"strconv"
	"strings"
	"time"

	"romanctl/internal/platform"
)

const termuxBin = "/data/data/com.termux/files/usr/bin"

// Battery é a leitura da bateria. Corrente em ampères, sempre positiva.
type Battery struct {
	Source         string  `json:"source"`
	Percent        float64 `json:"percent"`
	Status         string  `json:"status"` // discharging, charging, full ou unknown
	Temperature    float64 `json:"temperatureC,omitempty"`
	Voltage        float64 `json:"voltageV,omitempty"`
	Current        float64 `json:"currentA,omitempty"`
	Power          float64 `json:"powerW,omitempty"`
	ChargeAh       float64 `json:"chargeAh,omitempty"`
	FullChargeAh   float64 `json:"fullChargeAh,omitempty"`
	HoursLeft      float64 `json:"hoursLeft,omitempty"`      // até acabar ou até encher
	HoursBasis     string  `json:"hoursBasis,omitempty"`     // como HoursLeft foi calculado
	RatePerHour    float64 `json:"percentPerHour,omitempty"` // queda medida pelo histórico
	RateWindowMins float64 `json:"rateWindowMinutes,omitempty"`
}

// ErrNoBattery indica uma máquina sem bateria, como um desktop.
var ErrNoBattery = i18n.Errorf("esta máquina não tem bateria")

// ReadBattery lê a fonte certa para a plataforma.
func ReadBattery() (*Battery, error) {
	p := platform.Current
	var b *Battery
	var err error
	switch {
	case p.OS == "darwin":
		b, err = readMacBattery()
	case p.Proot || fileExists(termuxBin+"/termux-battery-status"):
		b, err = readTermuxBattery()
	default:
		b, err = readSysfsBattery("/sys/class/power_supply")
	}
	if err != nil {
		return nil, err
	}
	b.finish()
	return b, nil
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

// finish calcula o que deriva da leitura.
func (b *Battery) finish() {
	if b.Power == 0 && b.Voltage > 0 && b.Current > 0 {
		b.Power = b.Voltage * b.Current
	}
	if b.Current <= 0 {
		return
	}
	switch b.Status {
	case "discharging":
		if b.ChargeAh > 0 {
			b.HoursLeft, b.HoursBasis = b.ChargeAh/b.Current, i18n.T("carga restante ÷ corrente atual")
		}
	case "charging":
		full := b.FullChargeAh
		if full == 0 && b.ChargeAh > 0 && b.Percent > 0 {
			full = b.ChargeAh / b.Percent * 100
		}
		if full > b.ChargeAh {
			b.HoursLeft, b.HoursBasis = (full-b.ChargeAh)/b.Current, i18n.T("carga que falta ÷ corrente atual")
		}
	}
}

func normalizeStatus(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "discharging", "not charging":
		return "discharging"
	case "charging":
		return "charging"
	case "full", "charged", "finishing charge", "ac attached":
		return "full"
	}
	return "unknown"
}

// ---------------------------------------------------------------- Termux

func readTermuxBattery() (*Battery, error) {
	bin := termuxBin + "/termux-battery-status"
	if !fileExists(bin) {
		return nil, i18n.Errorf("instale o app Termux:API e, no Termux, rode: pkg install termux-api")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Env = append(os.Environ(), "PATH="+termuxBin+":"+os.Getenv("PATH"),
		"PREFIX=/data/data/com.termux/files/usr", "LD_LIBRARY_PATH=")
	out, err := cmd.Output()
	if err != nil {
		return nil, i18n.Errorf("termux-battery-status falhou (%v); o app Termux:API está instalado?", err)
	}
	return parseTermuxBattery(out)
}

func parseTermuxBattery(data []byte) (*Battery, error) {
	var raw struct {
		Percentage    float64  `json:"percentage"`
		Status        string   `json:"status"`
		Temperature   float64  `json:"temperature"`
		Voltage       float64  `json:"voltage"`
		Current       float64  `json:"current"`
		CurrentAvg    *float64 `json:"current_average"`
		ChargeCounter float64  `json:"charge_counter"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, i18n.Errorf("resposta inesperada do termux-battery-status: %w", err)
	}
	current := raw.Current
	if raw.CurrentAvg != nil && *raw.CurrentAvg != 0 {
		current = *raw.CurrentAvg // varia menos que a leitura instantânea
	}
	b := &Battery{Source: "termux-api", Percent: raw.Percentage, Status: normalizeStatus(raw.Status),
		Temperature: raw.Temperature, Voltage: raw.Voltage / 1000, Current: scaleToAmps(current)}
	// charge_counter vem em µAh, mas há aparelhos que usam mAh.
	switch counter := raw.ChargeCounter; {
	case counter > 200000:
		b.ChargeAh = counter / 1e6
	case counter > 0:
		b.ChargeAh = counter / 1e3
	}
	return b, nil
}

// O Android devolve microampères, mas alguns fabricantes usam miliampères.
func scaleToAmps(v float64) float64 {
	if v < 0 {
		v = -v
	}
	if v >= 20000 {
		return v / 1e6
	}
	return v / 1e3
}

// ---------------------------------------------------------------- Linux

func readSysfsBattery(root string) (*Battery, error) {
	dirs, _ := filepath.Glob(filepath.Join(root, "*"))
	for _, dir := range dirs {
		kind, err := os.ReadFile(filepath.Join(dir, "type"))
		if err != nil || strings.TrimSpace(string(kind)) != "Battery" {
			continue
		}
		num := func(name string) float64 {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				return 0
			}
			v, _ := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64)
			if v < 0 {
				return -v
			}
			return v
		}
		status, _ := os.ReadFile(filepath.Join(dir, "status"))
		b := &Battery{Source: "sysfs", Percent: num("capacity"), Status: normalizeStatus(string(status)),
			Voltage: num("voltage_now") / 1e6, Current: num("current_now") / 1e6, Power: num("power_now") / 1e6,
			ChargeAh: num("charge_now") / 1e6, FullChargeAh: num("charge_full") / 1e6}
		// Algumas baterias informam energia (Wh) em vez de carga (Ah).
		if b.Voltage > 0 {
			if b.Current == 0 && b.Power > 0 {
				b.Current = b.Power / b.Voltage
			}
			if b.ChargeAh == 0 {
				b.ChargeAh = num("energy_now") / 1e6 / b.Voltage
			}
			if b.FullChargeAh == 0 {
				b.FullChargeAh = num("energy_full") / 1e6 / b.Voltage
			}
		}
		return b, nil
	}
	return nil, ErrNoBattery
}

// ---------------------------------------------------------------- macOS

var (
	pmsetLine = regexp.MustCompile(`(\d+)%;\s*([^;]+);\s*(?:(\d+):(\d+)\s+remaining)?`)
	ioregNum  = regexp.MustCompile(`"(\w+)"\s*=\s*(-?\d+)`)
)

func readMacBattery() (*Battery, error) {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return nil, i18n.Errorf("pmset falhou: %w", err)
	}
	b, err := parsePmset(string(out))
	if err != nil {
		return nil, err
	}
	if reg, err := exec.Command("ioreg", "-rn", "AppleSmartBattery").Output(); err == nil {
		applyIoreg(b, string(reg))
	}
	return b, nil
}

func parsePmset(out string) (*Battery, error) {
	m := pmsetLine.FindStringSubmatch(out)
	if m == nil {
		return nil, ErrNoBattery
	}
	percent, _ := strconv.ParseFloat(m[1], 64)
	b := &Battery{Source: "pmset", Percent: percent, Status: normalizeStatus(m[2])}
	if m[3] != "" {
		h, _ := strconv.ParseFloat(m[3], 64)
		min, _ := strconv.ParseFloat(m[4], 64)
		b.HoursLeft, b.HoursBasis = h+min/60, i18n.T("estimativa do macOS")
	}
	return b, nil
}

func applyIoreg(b *Battery, out string) {
	values := map[string]float64{}
	for _, m := range ioregNum.FindAllStringSubmatch(out, -1) {
		v, _ := strconv.ParseFloat(m[2], 64)
		values[m[1]] = v
	}
	b.Voltage = values["Voltage"] / 1000
	amps := values["InstantAmperage"]
	if amps == 0 {
		amps = values["Amperage"]
	}
	if amps < 0 {
		amps = -amps
	}
	b.Current = amps / 1000
	b.Temperature = values["Temperature"] / 100
	b.ChargeAh = values["AppleRawCurrentCapacity"] / 1000
	b.FullChargeAh = values["AppleRawMaxCapacity"] / 1000
}
