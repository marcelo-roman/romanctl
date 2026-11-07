package sysinfo

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func near(a, b float64) bool { return math.Abs(a-b) < 0.01 }

func TestTermuxBateriaReal(t *testing.T) {
	b, err := parseTermuxBattery([]byte(`{"present":true,"plugged":"UNPLUGGED","status":"DISCHARGING","temperature":23.3,
		"voltage":4281,"current":-937000,"current_average":-885000,"percentage":90,"charge_counter":9151630,"energy":null}`))
	if err != nil {
		t.Fatal(err)
	}
	b.finish()
	if !near(b.Current, 0.885) || !near(b.Voltage, 4.281) || !near(b.ChargeAh, 9.15) || b.Status != "discharging" {
		t.Fatalf("%+v", b)
	}
	if !near(b.Power, 3.79) || !near(b.HoursLeft, 10.34) {
		t.Fatalf("potência %.2f, autonomia %.2f", b.Power, b.HoursLeft)
	}
}

func TestTermuxEmMiliamperes(t *testing.T) {
	b, err := parseTermuxBattery([]byte(`{"status":"CHARGING","voltage":4000,"current":1500,"percentage":50,"charge_counter":5000}`))
	if err != nil {
		t.Fatal(err)
	}
	b.finish()
	if !near(b.Current, 1.5) || !near(b.ChargeAh, 5) {
		t.Fatalf("%+v", b)
	}
	if !near(b.HoursLeft, 3.33) { // faltam 5 Ah a 1,5 A
		t.Fatalf("até encher: %.2f", b.HoursLeft)
	}
}

func TestSysfsEnergiaEmVezDeCarga(t *testing.T) {
	root := t.TempDir()
	write := func(dir, name, v string) {
		os.MkdirAll(filepath.Join(root, dir), 0o755)
		os.WriteFile(filepath.Join(root, dir, name), []byte(v+"\n"), 0o644)
	}
	write("AC", "type", "Mains")
	write("BAT0", "type", "Battery")
	write("BAT0", "capacity", "50")
	write("BAT0", "status", "Discharging")
	write("BAT0", "voltage_now", "12000000")
	write("BAT0", "power_now", "24000000")
	write("BAT0", "energy_now", "48000000")
	b, err := readSysfsBattery(root)
	if err != nil {
		t.Fatal(err)
	}
	b.finish()
	if !near(b.Power, 24) || !near(b.Current, 2) || !near(b.ChargeAh, 4) || !near(b.HoursLeft, 2) {
		t.Fatalf("%+v", b)
	}
}

func TestSemBateria(t *testing.T) {
	if _, err := readSysfsBattery(t.TempDir()); err != ErrNoBattery {
		t.Fatalf("erro: %v", err)
	}
}

func TestMacBateria(t *testing.T) {
	b, err := parsePmset("Now drawing from 'Battery Power'\n -InternalBattery-0 (id=1)\t87%; discharging; 5:30 remaining present: true\n")
	if err != nil {
		t.Fatal(err)
	}
	if b.Percent != 87 || b.Status != "discharging" || b.HoursLeft != 5.5 {
		t.Fatalf("%+v", b)
	}
	applyIoreg(b, `"Voltage" = 12500
  "InstantAmperage" = -1500
  "Temperature" = 3050
  "AppleRawCurrentCapacity" = 4000`)
	b.finish()
	if !near(b.Current, 1.5) || !near(b.Voltage, 12.5) || !near(b.Power, 18.75) || !near(b.Temperature, 30.5) {
		t.Fatalf("%+v", b)
	}
	if _, err := parsePmset("Now drawing from 'AC Power'\n"); err != ErrNoBattery {
		t.Fatalf("desktop: %v", err)
	}
}

func TestProcStat(t *testing.T) {
	total, idle, ok := parseProcStat("cpu  100 0 100 700 100 0 0 0 50 0\ncpu0 1 1\n")
	if !ok || total != 1000 || idle != 800 {
		t.Fatalf("%d %d %v", total, idle, ok)
	}
	if v, ok := cpuPercent(1000, 800, 2000, 1300); !ok || v != 50 {
		t.Fatalf("%v %v", v, ok)
	}
	if _, ok := cpuPercent(1000, 800, 1000, 800); ok {
		t.Fatal("sem intervalo não há como medir")
	}
}

func TestPidStatComEspacosNoNome(t *testing.T) {
	line := "42 (npm install (x)) S 1 1 1 0 -1 0 0 0 0 0 120 30 0 0 20 0 1 0 100 1000 77 0"
	ticks, rss, name, ok := parsePidStat(line)
	if !ok || ticks != 150 || rss != 77 || name != "npm install (x)" {
		t.Fatalf("%d %d %q %v", ticks, rss, name, ok)
	}
}

func TestDiffProcs(t *testing.T) {
	before := map[int]procTicks{1: {ticks: 100}, 2: {ticks: 500}}
	after := map[int]procTicks{1: {ticks: 150}, 2: {ticks: 500}, 3: {ticks: 9999}}
	cpu := map[int]float64{}
	for _, p := range diffProcs(before, after, time.Second, 100) {
		cpu[p.PID] = p.CPU
	}
	if !near(cpu[1], 50) || cpu[2] != 0 || cpu[3] != 0 {
		t.Fatalf("%v", cpu)
	}
}

func TestMemoria(t *testing.T) {
	m := parseMeminfo("MemTotal:  1000 kB\nMemAvailable: 400 kB\nSwapTotal: 200 kB\nSwapFree: 100 kB\n")
	if m.Total != 1024000 || m.Available != 409600 || m.SwapFree != 102400 {
		t.Fatalf("%+v", m)
	}
	got := parseVMStat("Mach Virtual Memory Statistics: (page size of 16384 bytes)\nPages free: 10.\nPages inactive: 20.\nPages wired down: 99.\n")
	if got != 30*16384 {
		t.Fatalf("%d", got)
	}
	if procs := parsePs("  12 3.5 2048 Google Chrome\n"); len(procs) != 1 || procs[0].Name != "Google Chrome" || procs[0].Memory != 2097152 {
		t.Fatalf("%+v", procs)
	}
}

func TestDischargeRate(t *testing.T) {
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	at := func(mins int, pct float64, status string) Reading {
		return Reading{At: base.Add(time.Duration(mins) * time.Minute), Percent: pct, Status: status}
	}
	rate, window := dischargeRate([]Reading{at(0, 90, "discharging"), at(30, 85, "discharging"), at(60, 80, "discharging")})
	if !near(rate, 10) || window != 60 {
		t.Fatalf("%v %v", rate, window)
	}
	// O carregador separa os trechos: só vale a descarga depois dele.
	rate, _ = dischargeRate([]Reading{at(0, 50, "discharging"), at(10, 60, "charging"), at(20, 70, "discharging"), at(50, 67, "discharging")})
	if !near(rate, 6) {
		t.Fatalf("%v", rate)
	}
	if rate, _ := dischargeRate([]Reading{at(0, 90, "discharging"), at(5, 89, "discharging")}); rate != 0 {
		t.Fatalf("5 minutos não bastam: %v", rate)
	}
	if rate, _ := dischargeRate([]Reading{at(0, 50, "discharging"), at(60, 60, "charging")}); rate != 0 {
		t.Fatalf("carregando não tem queda: %v", rate)
	}
}

func samples(f func(i int) (cores, watts float64)) []Sample {
	var out []Sample
	for i := 0; i < 20; i++ {
		c, w := f(i)
		out = append(out, Sample{Cores: c, Power: w})
	}
	return out
}

func TestEstimateRecuperaRelacaoConhecida(t *testing.T) {
	// 2 W de base e 1,5 W por núcleo.
	fit := Estimate(samples(func(i int) (float64, float64) {
		c := float64(i%4) * 0.5
		return c, 2 + 1.5*c
	}))
	if !fit.Reliable || !near(fit.WattsCore, 1.5) || !near(fit.BaseWatts, 2) || !near(fit.R2, 1) {
		t.Fatalf("%+v", fit)
	}
	if !near(fit.OwnWatts, 1.5*fit.MeanCores) {
		t.Fatalf("%+v", fit)
	}
}

func TestEstimateRecusaDadosRuins(t *testing.T) {
	cases := map[string][]Sample{
		"amostras": samples(func(int) (float64, float64) { return 1, 3 })[:5],
		"variou":   samples(func(i int) (float64, float64) { return 1, 3 + float64(i%3) }),
		"acompanhou": samples(func(i int) (float64, float64) {
			return float64(i%2) * 2, 3 + float64(i%5)
		}),
	}
	for want, data := range cases {
		fit := Estimate(data)
		if fit.Reliable || !strings.Contains(fit.Unreliable, want) || fit.OwnWatts != 0 {
			t.Errorf("%s: %+v", want, fit)
		}
	}
}
