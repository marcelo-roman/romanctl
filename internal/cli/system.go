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
		sub, err = ui.Choose(i18n.T("System"),
			ui.Opt(i18n.T("CPU, memory, disk, battery and processes"), "status"),
			ui.Opt(i18n.T("Battery and time left"), "battery"),
			ui.Opt(i18n.T("Measure power over time"), "watch"))
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
	return ui.Usagef("usage: romanctl system [status|battery|watch] [--seconds N] [--interval N]")
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
	ui.Spin(i18n.T("Measuring CPU, memory and battery..."), func(func(string)) { s = sysinfo.Take(time.Second) })
	if ui.JSONOutput {
		return ui.PrintJSON(s)
	}
	i18n.Printf("%s %s\n\n", ui.STitle.Render(i18n.T("System")), s.Platform.Describe())

	fmt.Println(ui.STitle.Render("CPU"))
	if s.SystemCPU != nil {
		i18n.Printf("  %-14s %.0f%% of %d cores\n", "Total", *s.SystemCPU, s.Cores)
	} else if s.Platform.Proot {
		i18n.Printf("  %-14s %s\n", "Total", ui.SGray.Render(i18n.T("proot does not report the device's total usage")))
	}
	label := i18n.T("Your processes")
	if s.Platform.Proot {
		label = "Ubuntu+Termux"
	}
	i18n.Printf("  %-14s %.2f cores, %s of memory\n", label, s.OwnCPUCores, gib(s.OwnMemory))

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Memory")))
	i18n.Printf("  %-14s %s available of %s\n", "RAM", gib(s.Memory.Available), gib(s.Memory.Total))
	if s.Memory.SwapTotal > 0 {
		i18n.Printf("  %-14s %s free of %s\n", "Swap", gib(s.Memory.SwapFree), gib(s.Memory.SwapTotal))
	}

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Disk")))
	i18n.Printf("  %-14s %s free of %s\n", s.Disk.Path, gib(s.Disk.Free), gib(s.Disk.Size))

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Battery")))
	printBattery(s.Battery, s.BatteryError)

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Top processes")))
	fmt.Println(ui.SGray.Render(i18n.Sprintf("  %-7s %6s %9s  %s", "PID", "CPU", i18n.T("MEMORY"), i18n.T("NAME"))))
	for i, p := range s.Top {
		if i == 8 {
			break
		}
		i18n.Printf("  %-7d %5.1f%% %6.0f MB  %s\n", p.PID, p.CPU, float64(p.Memory)/(1<<20), p.Name)
	}
	if s.Platform.Proot {
		fmt.Println(ui.SGray.Render(i18n.T("\n  Android only shows this app's processes. To measure what Ubuntu uses: romanctl system watch")))
	}
	return nil
}

var batteryStatus = map[string]string{"discharging": i18n.T("discharging"), "charging": i18n.T("charging"), "full": i18n.T("fully charged"), "unknown": i18n.T("unknown")}

func printBattery(b *sysinfo.Battery, errText string) {
	if b == nil {
		fmt.Println("  " + ui.SGray.Render(errText))
		return
	}
	i18n.Printf("  %-14s %.0f%%, %s\n", i18n.T("Level"), b.Percent, batteryStatus[b.Status])
	if b.Power > 0 {
		i18n.Printf("  %-14s %.1f W (%.2f A at %.2f V)\n", i18n.T("Total power"), b.Power, b.Current, b.Voltage)
	}
	if b.Temperature > 0 {
		i18n.Printf("  %-14s %.1f °C\n", i18n.T("Temperature"), b.Temperature)
	}
	if b.ChargeAh > 0 {
		i18n.Printf("  %-14s %.2f Ah\n", i18n.T("Charge left"), b.ChargeAh)
	}
	if b.HoursLeft > 0 {
		label := i18n.T("Time left")
		if b.Status == "charging" {
			label = i18n.T("Until full")
		}
		i18n.Printf("  %-14s %s %s\n", label, hours(b.HoursLeft), ui.SGray.Render("("+b.HoursBasis+")"))
	}
	if b.RatePerHour > 0 {
		i18n.Printf("  %-14s %.1f%%/h over the last %.0f min, or %s until empty\n", i18n.T("Measured drain"), b.RatePerHour, b.RateWindowMins, hours(b.Percent/b.RatePerHour))
	}
}

func systemBattery() error {
	var b *sysinfo.Battery
	var err error
	ui.Spin(i18n.T("Reading the battery..."), func(func(string)) { b, err = sysinfo.ReadBattery() })
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
		return 0, ui.Usagef("--%s needs a positive number", name)
	}
	return v, nil
}

// systemWatch measures power draw and CPU together to estimate what the visible processes consume.
func systemWatch() error {
	defSeconds := 0
	if ui.JSONOutput {
		defSeconds = 120 // without a terminal there is no Ctrl+C to stop it
	}
	seconds, err := intFlag("seconds", defSeconds)
	if err != nil {
		return err
	}
	if _, given := flagVals["seconds"]; !given && ui.Interactive() {
		choice, err := ui.Choose(i18n.T("Measure for how long?"),
			ui.Opt(i18n.T("Until I press Ctrl+C"), "0"),
			ui.Opt(i18n.T("1 minute"), "60"),
			ui.Opt(i18n.T("5 minutes"), "300"),
			ui.Opt(i18n.T("10 minutes (best for an estimate)"), "600"),
			ui.Opt(i18n.T("30 minutes"), "1800"))
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
		ui.Warnf("The battery is not discharging: unplug the charger for the measurement to make sense.")
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	defer signal.Stop(stop)

	if !ui.JSONOutput {
		ui.Infof("Measuring every %ds. Ctrl+C stops and shows the result.", interval)
		fmt.Println(ui.SGray.Render(i18n.Sprintf("  %-8s %6s %8s %8s %9s", i18n.T("TIME"), i18n.T("LEVEL"), i18n.T("POWER"), i18n.T("CORES"), i18n.T("CHARGE"))))
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
	own := i18n.T("Your processes")
	if platform.Current.Proot {
		own = "Ubuntu+Termux"
	}
	fmt.Println("\n" + ui.STitle.Render(i18n.T("Result")))
	i18n.Printf("  %-22s %s, %d samples\n", i18n.T("Duration"), elapsed.Round(time.Second), fit.Samples)
	i18n.Printf("  %-22s %.1f W\n", i18n.T("Average total power"), fit.MeanPower)
	i18n.Printf("  %-22s %.2f cores\n", own+" (CPU)", fit.MeanCores)
	if usedAh > 0 && elapsed > time.Minute {
		i18n.Printf("  %-22s %.3f Ah, or %.2f A on average\n", i18n.T("Charge used (measured)"), usedAh, usedAh/elapsed.Hours())
	}
	if fit.Reliable {
		i18n.Printf("  %-22s about %.1f W, %.0f%% of the total %s\n", own+i18n.T(" (estimated)"), fit.OwnWatts, fit.OwnShare, ui.SGray.Render(i18n.Sprintf("(%.2f W per core, R²=%.2f)", fit.WattsCore, fit.R2)))
		fmt.Println(ui.SGray.Render(i18n.T("  Statistical estimate: it only considers the CPU, not the network or memory.")))
		return nil
	}
	ui.Warnf("No estimate in watts: %s.", fit.Unreliable)
	return nil
}
