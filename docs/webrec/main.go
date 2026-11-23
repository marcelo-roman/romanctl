package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type frame struct {
	path string
	ts   float64
}

var (
	page   *rod.Page
	frames []frame
	outDir string
)

const cursorJS = `() => {
  if (document.getElementById('demo-cursor')) return;
  const c = document.createElement('div');
  c.id = 'demo-cursor';
  c.style.cssText = 'position:fixed;left:720px;top:450px;width:22px;height:22px;margin:-11px 0 0 -11px;border-radius:50%;background:rgba(236,72,153,.35);border:2px solid #ec4899;z-index:2147483647;pointer-events:none;transition:left .6s ease,top .6s ease,transform .15s;';
  document.body.appendChild(c);
}`

func cursor() { page.MustEval(cursorJS) }

func pause(d time.Duration) { time.Sleep(d) }

func moveTo(el *rod.Element) {
	cursor()
	el.MustScrollIntoView()
	pause(250 * time.Millisecond)
	box := el.MustShape().Box()
	page.MustEval(`(x, y) => { const c = document.getElementById('demo-cursor'); c.style.left = x + 'px'; c.style.top = y + 'px'; }`, box.X+box.Width/2, box.Y+box.Height/2)
	pause(750 * time.Millisecond)
}

func pulse() {
	page.MustEval(`() => { const c = document.getElementById('demo-cursor'); c.style.transform = 'scale(.6)'; setTimeout(() => c.style.transform = 'scale(1)', 150); }`)
}

func find(sel, re string) *rod.Element {
	return page.Timeout(60 * time.Second).MustElementR(sel, re).CancelTimeout()
}

func click(el *rod.Element) {
	moveTo(el)
	pulse()
	el.MustClick()
	pause(600 * time.Millisecond)
}

func clickText(sel, re string) { click(find(sel, re)) }

func xpath(x string) *rod.Element { return page.Timeout(60 * time.Second).MustElementX(x).CancelTimeout() }

func typeSlow(el *rod.Element, text string) {
	click(el)
	el.MustSelectAllText().MustInput("")
	for _, r := range text {
		el.MustInput(string(r))
		pause(90 * time.Millisecond)
	}
	pause(400 * time.Millisecond)
}

func waitFor(re string, timeout time.Duration) {
	page.Timeout(timeout).MustElementR("body *", re)
}

func smoothScroll(px int) {
	page.MustEval(`(y) => window.scrollBy({top: y, behavior: 'smooth'})`, px)
	pause(1500 * time.Millisecond)
}

func navClick(name string) {
	clickText("nav a", "^"+regexp.QuoteMeta(name)+"$")
	pause(1500 * time.Millisecond)
}

func main() {
	url, out := os.Args[1], os.Args[2]
	outDir = filepath.Join(out, "frames")
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0o755)

	l := launcher.New().Bin("/usr/bin/google-chrome").Headless(true).NoSandbox(true).
		Set("headless", "new").Set("disable-gpu").Set("disable-software-rasterizer").
		Set("window-size", "1440,900").Set("force-color-profile", "srgb")
	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	defer browser.MustClose()

	page = browser.MustPage("")
	page.MustSetViewport(1440, 900, 1, false)
	_ = proto.EmulationSetEmulatedMedia{Features: []*proto.EmulationMediaFeature{{Name: "prefers-color-scheme", Value: "dark"}}}.Call(page)

	go page.EachEvent(func(e *proto.PageScreencastFrame) {
		p := filepath.Join(outDir, fmt.Sprintf("%06d.jpg", len(frames)))
		_ = os.WriteFile(p, e.Data, 0o644)
		frames = append(frames, frame{p, float64(e.Metadata.Timestamp)})
		_ = proto.PageScreencastFrameAck{SessionID: e.SessionID}.Call(page)
	})()
	q := 85
	if err := (proto.PageStartScreencast{Format: proto.PageStartScreencastFormatJpeg, Quality: &q, MaxWidth: intp(1440), MaxHeight: intp(900), EveryNthFrame: intp(1)}).Call(page); err != nil {
		panic(err)
	}

	page.MustNavigate(url).MustWaitLoad()
	cursor()
	run()
	pause(2 * time.Second)
	proto.PageStopScreencast{}.Call(page)
	pause(500 * time.Millisecond)
	writeConcat(out)
}

func intp(i int) *int { return &i }

func writeConcat(out string) {
	sort.Slice(frames, func(i, j int) bool { return frames[i].ts < frames[j].ts })
	var b strings.Builder
	for i, f := range frames {
		d := 1.0 / 25
		if i+1 < len(frames) {
			d = frames[i+1].ts - f.ts
		}
		fmt.Fprintf(&b, "file '%s'\nduration %.4f\n", f.path, d)
	}
	if len(frames) > 0 {
		fmt.Fprintf(&b, "file '%s'\n", frames[len(frames)-1].path)
	}
	os.WriteFile(filepath.Join(out, "concat.txt"), []byte(b.String()), 0o644)
	fmt.Println("frames:", len(frames))
}
