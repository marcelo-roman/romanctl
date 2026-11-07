package sysinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"romanctl/internal/i18n"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"romanctl/internal/platform"
)

// Proc é um processo com o uso medido durante a amostra.
type Proc struct {
	PID    int     `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpuPercentOfOneCore"`
	Memory int64   `json:"rssBytes"`
}

// Memory é o uso de memória da máquina.
type Memory struct {
	Total     int64 `json:"totalBytes"`
	Available int64 `json:"availableBytes"`
	SwapTotal int64 `json:"swapTotalBytes"`
	SwapFree  int64 `json:"swapFreeBytes"`
}

// Disk é o espaço da partição da pasta do usuário.
type Disk struct {
	Path string `json:"path"`
	Size int64  `json:"sizeBytes"`
	Free int64  `json:"freeBytes"`
}

// Snapshot é o retrato da máquina.
type Snapshot struct {
	Platform     platform.Info `json:"platform"`
	Cores        int           `json:"cores"`
	SystemCPU    *float64      `json:"systemCpuPercent"` // nulo quando a máquina não informa
	OwnCPUCores  float64       `json:"visibleProcessesCores"`
	OwnMemory    int64         `json:"visibleProcessesRssBytes"`
	Memory       Memory        `json:"memory"`
	Disk         Disk          `json:"disk"`
	Battery      *Battery      `json:"battery,omitempty"`
	BatteryError string        `json:"batteryError,omitempty"`
	Top          []Proc        `json:"top"`
}

// Take mede por um intervalo, porque uso de CPU só existe entre duas leituras.
func Take(interval time.Duration) Snapshot {
	s := Snapshot{Platform: platform.Current, Cores: runtime.NumCPU()}
	s.Top, s.SystemCPU = sampleCPU(interval)
	for _, p := range s.Top {
		s.OwnCPUCores += p.CPU / 100
		s.OwnMemory += p.Memory
	}
	sort.Slice(s.Top, func(i, j int) bool {
		if s.Top[i].CPU != s.Top[j].CPU {
			return s.Top[i].CPU > s.Top[j].CPU
		}
		return s.Top[i].Memory > s.Top[j].Memory
	})
	s.Memory = readMemory()
	s.Disk = readDisk()
	if b, err := ReadBattery(); err != nil {
		s.BatteryError = err.Error()
	} else {
		RecordAndRate(b, time.Now())
		s.Battery = b
	}
	return s
}

// ---------------------------------------------------------------- CPU

type procTicks struct {
	ticks int64
	rss   int64
	name  string
}

func sampleCPU(interval time.Duration) ([]Proc, *float64) {
	if platform.Current.OS == "darwin" {
		return sampleMac()
	}
	beforeTotal, beforeIdle, statOK := readProcStat()
	before := readProcs()
	time.Sleep(interval)
	after := readProcs()
	afterTotal, afterIdle, _ := readProcStat()

	var system *float64
	// No proot o /proc/stat é simulado: um número dali seria inventado.
	if statOK && !platform.Current.Proot {
		if v, ok := cpuPercent(beforeTotal, beforeIdle, afterTotal, afterIdle); ok {
			system = &v
		}
	}
	return diffProcs(before, after, interval, 100), system
}

func cpuPercent(total1, idle1, total2, idle2 int64) (float64, bool) {
	dt := total2 - total1
	if dt <= 0 {
		return 0, false
	}
	return 100 * float64(dt-(idle2-idle1)) / float64(dt), true
}

func readProcStat() (total, idle int64, ok bool) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, false
	}
	return parseProcStat(string(b))
}

func parseProcStat(text string) (total, idle int64, ok bool) {
	line, _, _ := strings.Cut(text, "\n")
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, false
	}
	for i, f := range fields[1:] {
		v, _ := strconv.ParseInt(f, 10, 64)
		if i < 8 { // guest já está contado em user
			total += v
		}
		if i == 3 || i == 4 { // idle e iowait
			idle += v
		}
	}
	return total, idle, true
}

func readProcs() map[int]procTicks {
	out := map[int]procTicks{}
	self := os.Getpid()
	dirs, _ := filepath.Glob("/proc/[0-9]*")
	for _, dir := range dirs {
		pid, err := strconv.Atoi(filepath.Base(dir))
		if err != nil || pid == self {
			continue
		}
		stat, err := os.ReadFile(dir + "/stat")
		if err != nil {
			continue
		}
		ticks, rss, name, ok := parsePidStat(string(stat))
		if !ok {
			continue
		}
		if cmd, err := os.ReadFile(dir + "/cmdline"); err == nil && len(cmd) > 0 {
			name = commandName(string(cmd), name)
		}
		out[pid] = procTicks{ticks, rss * int64(os.Getpagesize()), name}
	}
	return out
}

// O nome do processo pode ter espaços e parênteses, então a linha é lida a partir do último ")".
func parsePidStat(text string) (ticks, rssPages int64, name string, ok bool) {
	open, closeIdx := strings.Index(text, "("), strings.LastIndex(text, ")")
	if open < 0 || closeIdx < open {
		return 0, 0, "", false
	}
	fields := strings.Fields(text[closeIdx+1:])
	if len(fields) < 22 {
		return 0, 0, "", false
	}
	utime, _ := strconv.ParseInt(fields[11], 10, 64)
	stime, _ := strconv.ParseInt(fields[12], 10, 64)
	rssPages, _ = strconv.ParseInt(fields[21], 10, 64)
	return utime + stime, rssPages, text[open+1 : closeIdx], true
}

func commandName(cmdline, fallback string) string {
	arg0, _, _ := strings.Cut(cmdline, "\x00")
	arg0 = filepath.Base(strings.TrimSpace(arg0))
	if arg0 == "" || arg0 == "." {
		return fallback
	}
	if len(arg0) > 28 {
		return arg0[:28]
	}
	return arg0
}

// hz é a unidade de /proc/PID/stat, 100 em praticamente todo Linux.
func diffProcs(before, after map[int]procTicks, interval time.Duration, hz float64) []Proc {
	var out []Proc
	for pid, now := range after {
		used := now.ticks
		if prev, ok := before[pid]; ok {
			used -= prev.ticks
		} else {
			used = 0 // nasceu durante a amostra: o total acumulado não é desse intervalo
		}
		if used < 0 {
			used = 0
		}
		out = append(out, Proc{PID: pid, Name: now.name, Memory: now.rss,
			CPU: float64(used) / hz / interval.Seconds() * 100})
	}
	return out
}

var macCPU = regexp.MustCompile(`CPU usage:\s*([\d.]+)% user,\s*([\d.]+)% sys`)

func sampleMac() ([]Proc, *float64) {
	var system *float64
	if out, err := exec.Command("top", "-l", "2", "-n", "0", "-s", "1").Output(); err == nil {
		// A primeira amostra do top é a média desde o boot: vale a última.
		if all := macCPU.FindAllStringSubmatch(string(out), -1); len(all) > 0 {
			m := all[len(all)-1]
			user, _ := strconv.ParseFloat(m[1], 64)
			sys, _ := strconv.ParseFloat(m[2], 64)
			v := user + sys
			system = &v
		}
	}
	out, err := exec.Command("ps", "-Aceo", "pid=,pcpu=,rss=,comm=").Output()
	if err != nil {
		return nil, system
	}
	return parsePs(string(out)), system
}

func parsePs(text string) []Proc {
	var procs []Proc
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pid, _ := strconv.Atoi(fields[0])
		cpu, _ := strconv.ParseFloat(fields[1], 64)
		rss, _ := strconv.ParseInt(fields[2], 10, 64)
		procs = append(procs, Proc{PID: pid, Name: strings.Join(fields[3:], " "), CPU: cpu, Memory: rss * 1024})
	}
	return procs
}

// ---------------------------------------------------------------- memória e disco

func readMemory() Memory {
	if platform.Current.OS == "darwin" {
		return readMacMemory()
	}
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return Memory{}
	}
	return parseMeminfo(string(b))
}

func parseMeminfo(text string) Memory {
	values := map[string]int64{}
	for _, line := range strings.Split(text, "\n") {
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			v, _ := strconv.ParseInt(fields[0], 10, 64)
			values[key] = v * 1024
		}
	}
	return Memory{values["MemTotal"], values["MemAvailable"], values["SwapTotal"], values["SwapFree"]}
}

func readMacMemory() Memory {
	var m Memory
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		m.Total, _ = strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	}
	if out, err := exec.Command("vm_stat").Output(); err == nil {
		m.Available = parseVMStat(string(out))
	}
	return m
}

var vmStatPageSize = regexp.MustCompile(`page size of (\d+) bytes`)

func parseVMStat(text string) int64 {
	pageSize := int64(4096)
	if m := vmStatPageSize.FindStringSubmatch(text); m != nil {
		pageSize, _ = strconv.ParseInt(m[1], 10, 64)
	}
	var pages int64
	for _, line := range strings.Split(text, "\n") {
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "Pages free", "Pages inactive", "Pages purgeable", "Pages speculative":
			v, _ := strconv.ParseInt(strings.TrimSuffix(strings.TrimSpace(rest), "."), 10, 64)
			pages += v
		}
	}
	return pages * pageSize
}

func readDisk() Disk {
	home, err := os.UserHomeDir()
	if err != nil {
		return Disk{}
	}
	var st syscall.Statfs_t
	if err := syscall.Statfs(home, &st); err != nil {
		return Disk{Path: home}
	}
	block := int64(st.Bsize)
	return Disk{Path: home, Size: int64(st.Blocks) * block, Free: int64(st.Bavail) * block}
}

// ---------------------------------------------------------------- watch

// Sample é uma medição do comando watch.
type Sample struct {
	At    time.Time `json:"at"`
	Power float64   `json:"powerW"`
	Cores float64   `json:"cores"`
}

// Fit é o resultado da regressão entre uso de CPU e consumo.
type Fit struct {
	Samples    int     `json:"samples"`
	MeanPower  float64 `json:"meanPowerW"`
	MeanCores  float64 `json:"meanCores"`
	WattsCore  float64 `json:"wattsPerCore,omitempty"`
	BaseWatts  float64 `json:"baselineW,omitempty"`
	OwnWatts   float64 `json:"estimatedW,omitempty"`
	OwnShare   float64 `json:"estimatedSharePercent,omitempty"`
	R2         float64 `json:"r2,omitempty"`
	Reliable   bool    `json:"reliable"`
	Unreliable string  `json:"unreliableReason,omitempty"`
}

// Estimate relaciona CPU e consumo para atribuir parte dos watts aos processos visíveis.
// A regressão só é aceita com amostras suficientes, variação de CPU e correlação visível.
func Estimate(samples []Sample) Fit {
	fit := Fit{Samples: len(samples)}
	if len(samples) == 0 {
		fit.Unreliable = i18n.T("nenhuma amostra")
		return fit
	}
	var sumX, sumY float64
	for _, s := range samples {
		sumX += s.Cores
		sumY += s.Power
	}
	n := float64(len(samples))
	fit.MeanCores, fit.MeanPower = sumX/n, sumY/n
	if len(samples) < 12 {
		fit.Unreliable = i18n.Sprintf("poucas amostras (%d de 12)", len(samples))
		return fit
	}
	var sxx, sxy, syy float64
	for _, s := range samples {
		dx, dy := s.Cores-fit.MeanCores, s.Power-fit.MeanPower
		sxx += dx * dx
		sxy += dx * dy
		syy += dy * dy
	}
	if sxx/n < 0.01 { // desvio padrão abaixo de 0,1 núcleo
		fit.Unreliable = i18n.T("o uso de CPU quase não variou; rode algo pesado durante parte da medição")
		return fit
	}
	if syy == 0 {
		fit.Unreliable = i18n.T("o consumo não variou")
		return fit
	}
	slope := sxy / sxx
	fit.R2 = sxy * sxy / (sxx * syy)
	if slope <= 0 || fit.R2 < 0.3 {
		fit.Unreliable = i18n.Sprintf("o consumo não acompanhou o uso de CPU (R²=%.2f); outros apps ou a tela pesaram mais", fit.R2)
		return fit
	}
	fit.WattsCore = slope
	fit.BaseWatts = fit.MeanPower - slope*fit.MeanCores
	fit.OwnWatts = slope * fit.MeanCores
	fit.OwnShare = 100 * fit.OwnWatts / fit.MeanPower
	fit.Reliable = true
	return fit
}

// SampleOwnCores mede a CPU dos processos visíveis, em núcleos.
func SampleOwnCores(interval time.Duration) float64 {
	procs, _ := sampleCPU(interval)
	var total float64
	for _, p := range procs {
		total += p.CPU / 100
	}
	return total
}
