// Package sysinfo reads the machine's CPU, memory, disk, processes and battery.
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

// Battery is a battery reading. Current in amperes, always positive.
type Battery struct {
	Source         string  `json:"source"`
	Percent        float64 `json:"percent"`
	Status         string  `json:"status"` // discharging, charging, full or unknown
	Temperature    float64 `json:"temperatureC,omitempty"`
	Voltage        float64 `json:"voltageV,omitempty"`
	Current        float64 `json:"currentA,omitempty"`
	Power          float64 `json:"powerW,omitempty"`
	ChargeAh       float64 `json:"chargeAh,omitempty"`
	FullChargeAh   float64 `json:"fullChargeAh,omitempty"`
	HoursLeft      float64 `json:"hoursLeft,omitempty"`      // until empty or until full
	HoursBasis     string  `json:"hoursBasis,omitempty"`     // how HoursLeft was calculated
	RatePerHour    float64 `json:"percentPerHour,omitempty"` // drop measured from the history
	RateWindowMins float64 `json:"rateWindowMinutes,omitempty"`
}

// ErrNoBattery means a machine without a battery, like a desktop.
var ErrNoBattery = i18n.Errorf("this machine has no battery")

// ReadBattery reads the right source for the platform.
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

// finish computes what derives from the reading.
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
			b.HoursLeft, b.HoursBasis = b.ChargeAh/b.Current, i18n.T("charge left ÷ current draw")
		}
	case "charging":
		full := b.FullChargeAh
		if full == 0 && b.ChargeAh > 0 && b.Percent > 0 {
			full = b.ChargeAh / b.Percent * 100
		}
		if full > b.ChargeAh {
			b.HoursLeft, b.HoursBasis = (full-b.ChargeAh)/b.Current, i18n.T("missing charge ÷ current draw")
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
		return nil, i18n.Errorf("install the Termux:API app and, in Termux, run: pkg install termux-api")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Env = append(os.Environ(), "PATH="+termuxBin+":"+os.Getenv("PATH"),
		"PREFIX=/data/data/com.termux/files/usr", "LD_LIBRARY_PATH=")
	out, err := cmd.Output()
	if err != nil {
		return nil, i18n.Errorf("termux-battery-status failed (%v); is the Termux:API app installed?", err)
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
		return nil, i18n.Errorf("unexpected response from termux-battery-status: %w", err)
	}
	current := raw.Current
	if raw.CurrentAvg != nil && *raw.CurrentAvg != 0 {
		current = *raw.CurrentAvg // varies less than the instantaneous reading
	}
	b := &Battery{Source: "termux-api", Percent: raw.Percentage, Status: normalizeStatus(raw.Status),
		Temperature: raw.Temperature, Voltage: raw.Voltage / 1000, Current: scaleToAmps(current)}
	// charge_counter comes in µAh, but some devices use mAh.
	switch counter := raw.ChargeCounter; {
	case counter > 200000:
		b.ChargeAh = counter / 1e6
	case counter > 0:
		b.ChargeAh = counter / 1e3
	}
	return b, nil
}

// Android returns microamperes, but some manufacturers use milliamperes.
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
		// Some batteries report energy (Wh) instead of charge (Ah).
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
		return nil, i18n.Errorf("pmset failed: %w", err)
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
		b.HoursLeft, b.HoursBasis = h+min/60, i18n.T("macOS estimate")
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
