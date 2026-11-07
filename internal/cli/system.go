package cli

import (
	"fmt"
	"os"
	"os/signal"
	"romanctl/internal/i18n"
	"strconv"
	"time"

	"romanctl/internal/platform"
	"romanctl/internal/sysinfo"
	"romanctl/internal/ui"
)

func systemCommand(args []string) error {
	sub := arg(args, 0)
	if sub == "" && ui.Interactive() {
		var err error
		sub, err = ui.Choose(i18n.T("Sistema"),
			ui.Opt(i18n.T("CPU, memória, disco, bateria e processos"), "status"),
			ui.Opt(i18n.T("Bateria e previsão de autonomia"), "battery"),
			ui.Opt(i18n.T("Medir consumo por um tempo"), "watch"))
		if err != nil {
			return err
		}
	}
	switch sub {
	case "", "status":
		return systemStatus()
	case "battery", "bateria":
		return systemBattery()
	case "watch", "medir":
		return systemWatch()
	}
	return ui.Usagef("use: romanctl system [status|battery|watch] [--seconds N] [--interval N]")
}

func gib(n int64) string { return i18n.Sprintf("%.1f GB", float64(n)/(1<<30)) }

func hours(h float64) string {
	if h <= 0 {
		return "—"
	}
	return i18n.Sprintf("%dh%02d", int(h), int(h*60)%60)
}

func systemStatus() error {
	var s sysinfo.Snapshot
	ui.Spin(i18n.T("Medindo CPU, memória e bateria..."), func(func(string)) { s = sysinfo.Take(time.Second) })
	if ui.JSONOutput {
		return ui.PrintJSON(s)
	}
	i18n.Printf("%s %s\n\n", ui.STitle.Render(i18n.T("Sistema")), s.Platform.Describe())

	fmt.Println(ui.STitle.Render("CPU"))
	if s.SystemCPU != nil {
		i18n.Printf("  %-14s %.0f%% de %d núcleos\n", "Total", *s.SystemCPU, s.Cores)
	} else if s.Platform.Proot {
		i18n.Printf("  %-14s %s\n", "Total", ui.SGray.Render(i18n.T("o proot não informa o uso total do aparelho")))
	}
	label := i18n.T("Seus processos")
	if s.Platform.Proot {
		label = "Ubuntu+Termux"
	}
	i18n.Printf("  %-14s %.2f núcleos, %s de memória\n", label, s.OwnCPUCores, gib(s.OwnMemory))

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Memória")))
	i18n.Printf("  %-14s %s disponíveis de %s\n", "RAM", gib(s.Memory.Available), gib(s.Memory.Total))
	if s.Memory.SwapTotal > 0 {
		i18n.Printf("  %-14s %s livres de %s\n", "Swap", gib(s.Memory.SwapFree), gib(s.Memory.SwapTotal))
	}

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Disco")))
	i18n.Printf("  %-14s %s livres de %s\n", s.Disk.Path, gib(s.Disk.Free), gib(s.Disk.Size))

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Bateria")))
	printBattery(s.Battery, s.BatteryError)

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Processos que mais consomem")))
	fmt.Println(ui.SGray.Render(i18n.Sprintf("  %-7s %6s %9s  %s", "PID", "CPU", i18n.T("MEMÓRIA"), i18n.T("NOME"))))
	for i, p := range s.Top {
		if i == 8 {
			break
		}
		i18n.Printf("  %-7d %5.1f%% %6.0f MB  %s\n", p.PID, p.CPU, float64(p.Memory)/(1<<20), p.Name)
	}
	if s.Platform.Proot {
		fmt.Println(ui.SGray.Render(i18n.T("\n  O Android só mostra os processos deste app. Para medir o gasto do Ubuntu: romanctl system watch")))
	}
	return nil
}

var batteryStatus = map[string]string{"discharging": i18n.T("descarregando"), "charging": i18n.T("carregando"), "full": i18n.T("carga completa"), "unknown": i18n.T("desconhecido")}

func printBattery(b *sysinfo.Battery, errText string) {
	if b == nil {
		fmt.Println("  " + ui.SGray.Render(errText))
		return
	}
	i18n.Printf("  %-14s %.0f%%, %s\n", i18n.T("Nível"), b.Percent, batteryStatus[b.Status])
	if b.Power > 0 {
		i18n.Printf("  %-14s %.1f W (%.2f A a %.2f V)\n", i18n.T("Consumo total"), b.Power, b.Current, b.Voltage)
	}
	if b.Temperature > 0 {
		i18n.Printf("  %-14s %.1f °C\n", i18n.T("Temperatura"), b.Temperature)
	}
	if b.ChargeAh > 0 {
		i18n.Printf("  %-14s %.2f Ah\n", i18n.T("Carga restante"), b.ChargeAh)
	}
	if b.HoursLeft > 0 {
		label := i18n.T("Autonomia")
		if b.Status == "charging" {
			label = i18n.T("Até encher")
		}
		i18n.Printf("  %-14s %s %s\n", label, hours(b.HoursLeft), ui.SGray.Render("("+b.HoursBasis+")"))
	}
	if b.RatePerHour > 0 {
		i18n.Printf("  %-14s %.1f%%/h nos últimos %.0f min, ou %s até zerar\n", i18n.T("Queda medida"), b.RatePerHour, b.RateWindowMins, hours(b.Percent/b.RatePerHour))
	}
}

func systemBattery() error {
	var b *sysinfo.Battery
	var err error
	ui.Spin(i18n.T("Lendo a bateria..."), func(func(string)) { b, err = sysinfo.ReadBattery() })
	if err != nil {
		return err
	}
	sysinfo.RecordAndRate(b, time.Now())
	if ui.JSONOutput {
		return ui.PrintJSON(b)
	}
	printBattery(b, "")
	return nil
}

func intFlag(name string, def int) (int, error) {
	raw, ok := flagVals[name]
	if !ok {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, ui.Usagef("--%s precisa de um número positivo", name)
	}
	return v, nil
}

// systemWatch mede consumo e CPU juntos para estimar o gasto dos processos visíveis.
func systemWatch() error {
	defSeconds := 0
	if ui.JSONOutput {
		defSeconds = 120 // sem terminal não há Ctrl+C para encerrar
	}
	seconds, err := intFlag("seconds", defSeconds)
	if err != nil {
		return err
	}
	if _, given := flagVals["seconds"]; !given && ui.Interactive() {
		choice, err := ui.Choose(i18n.T("Medir por quanto tempo?"),
			ui.Opt(i18n.T("Até eu apertar Ctrl+C"), "0"),
			ui.Opt(i18n.T("1 minuto"), "60"),
			ui.Opt(i18n.T("5 minutos"), "300"),
			ui.Opt(i18n.T("10 minutos (melhor para estimar)"), "600"),
			ui.Opt(i18n.T("30 minutos"), "1800"))
		if err != nil {
			return err
		}
		seconds, _ = strconv.Atoi(choice)
	}
	interval, err := intFlag("interval", 5)
	if err != nil {
		return err
	}
	first, err := sysinfo.ReadBattery()
	if err != nil {
		return err
	}
	if first.Status != "discharging" && !ui.JSONOutput {
		ui.Warnf("A bateria não está descarregando: desconecte o carregador para a medição fazer sentido.")
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	defer signal.Stop(stop)

	if !ui.JSONOutput {
		ui.Infof("Medindo a cada %ds. Ctrl+C encerra e mostra o resultado.", interval)
		fmt.Println(ui.SGray.Render(i18n.Sprintf("  %-8s %6s %8s %8s %9s", i18n.T("HORA"), i18n.T("NÍVEL"), i18n.T("CONSUMO"), i18n.T("NÚCLEOS"), i18n.T("CARGA"))))
	}
	start := time.Now()
	var samples []sysinfo.Sample
	last := first
loop:
	for seconds == 0 || time.Since(start) < time.Duration(seconds)*time.Second {
		cores := sysinfo.SampleOwnCores(time.Duration(interval) * time.Second)
		b, err := sysinfo.ReadBattery()
		if err != nil {
			return err
		}
		sysinfo.RecordAndRate(b, time.Now())
		last = b
		samples = append(samples, sysinfo.Sample{At: time.Now(), Power: b.Power, Cores: cores})
		if !ui.JSONOutput {
			i18n.Printf("  %-8s %5.0f%% %6.1f W %8.2f %6.3f Ah\n", time.Now().Format("15:04:05"), b.Percent, b.Power, cores, b.ChargeAh)
		}
		select {
		case <-stop:
			break loop
		default:
		}
	}
	return printWatchResult(sysinfo.Estimate(samples), first, last, time.Since(start))
}

func printWatchResult(fit sysinfo.Fit, first, last *sysinfo.Battery, elapsed time.Duration) error {
	usedAh := first.ChargeAh - last.ChargeAh
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"fit": fit, "elapsedSeconds": elapsed.Seconds(), "chargeUsedAh": usedAh, "battery": last})
	}
	own := i18n.T("Seus processos")
	if platform.Current.Proot {
		own = "Ubuntu+Termux"
	}
	fmt.Println("\n" + ui.STitle.Render(i18n.T("Resultado")))
	i18n.Printf("  %-22s %s, %d amostras\n", i18n.T("Duração"), elapsed.Round(time.Second), fit.Samples)
	i18n.Printf("  %-22s %.1f W\n", i18n.T("Consumo total médio"), fit.MeanPower)
	i18n.Printf("  %-22s %.2f núcleos\n", own+" (CPU)", fit.MeanCores)
	if usedAh > 0 && elapsed > time.Minute {
		i18n.Printf("  %-22s %.3f Ah, ou %.2f A em média\n", i18n.T("Carga gasta (medida)"), usedAh, usedAh/elapsed.Hours())
	}
	if fit.Reliable {
		i18n.Printf("  %-22s cerca de %.1f W, %.0f%% do total %s\n", own+i18n.T(" (estimado)"), fit.OwnWatts, fit.OwnShare, ui.SGray.Render(i18n.Sprintf("(%.2f W por núcleo, R²=%.2f)", fit.WattsCore, fit.R2)))
		fmt.Println(ui.SGray.Render(i18n.T("  Estimativa estatística: considera só a CPU, não a rede nem a memória.")))
		return nil
	}
	ui.Warnf("Sem estimativa em watts: %s.", fit.Unreliable)
	return nil
}
