package main

import "time"

func run() {
	waitFor("Dashboard", 30*time.Second)
	pause(3 * time.Second)
	smoothScroll(500)
	pause(1500 * time.Millisecond)
	smoothScroll(-500)

	navClick("Projects")
	pause(1500 * time.Millisecond)
	clickText("a,button", "^New project$")
	pause(1500 * time.Millisecond)
	smoothScroll(400)
	clickText("button.card", "hexagonal-ddd[\\s\\S]*Go API")
	pause(1000 * time.Millisecond)
	click(xpath("//fieldset[legend[contains(., 'Services')]]//input[@type='checkbox']"))
	pause(800 * time.Millisecond)
	typeSlow(xpath("//label[contains(., 'Project name')]/following::input[1] | //label[contains(., 'Project name')]//input"), "shop")
	clickText("button", "^Create project$")
	waitFor("^succeeded$", 120*time.Second)
	pause(3 * time.Second)

	navClick("Projects")
	waitFor("^shop$", 30*time.Second)
	pause(1500 * time.Millisecond)
	clickText("button", "^Generate code$")
	pause(1000 * time.Millisecond)
	typeSlow(xpath("//label[contains(., 'Entity name')]/following::input[1] | //label[contains(., 'Entity name')]//input"), "Product")
	clickText("dialog button[type=submit], [role=dialog] button[type=submit]", ".")
	pause(6 * time.Second)

	navClick("Projects")
	clickText("button", "^Up$")
	waitFor("^Down$", 90*time.Second)
	pause(2 * time.Second)
	clickText("button", "^Logs$")
	pause(4 * time.Second)
	closeDialog()
	pause(1000 * time.Millisecond)

	navClick("Dashboard")
	pause(3 * time.Second)

	navClick("Projects")
	clickText("button", "^Down$")
	waitFor("^Up$", 60*time.Second)
	pause(1500 * time.Millisecond)

	navClick("Services")
	pause(3 * time.Second)
	smoothScroll(500)
	smoothScroll(-500)

	navClick("Tools")
	waitFor("of [0-9]+", 90*time.Second)
	pause(2 * time.Second)
	typeSlow(xpath("//input[@type='search'] | //label[contains(., 'Search tools')]/following::input[1]"), "helm")
	pause(1500 * time.Millisecond)
	clickText("button", "^Choose version$")
	pause(4 * time.Second)
	closeDialog()
	pause(1000 * time.Millisecond)

	navClick("Templates")
	pause(3 * time.Second)
	smoothScroll(600)
	smoothScroll(-600)

	navClick("Jobs")
	pause(2500 * time.Millisecond)
	clickText("main a", "Create|New|shop")
	pause(4 * time.Second)

	navClick("Dashboard")
	pause(1500 * time.Millisecond)
	lang := page.MustElement("header select")
	moveTo(lang)
	pulse()
	lang.MustSelect("PT")
	pause(4 * time.Second)
	moveTo(lang)
	pulse()
	lang.MustSelect("EN")
	pause(2 * time.Second)

	navClick("Projects")
	clickText("button", "^Remove$")
	pause(1000 * time.Millisecond)
	typeSlow(xpath("//*[@role='dialog' or self::dialog]//input"), "shop")
	clickText("dialog button, [role=dialog] button", "^Remove$")
	pause(4 * time.Second)
}

func closeDialog() {
	click(xpath("//*[@role='dialog' or self::dialog]//button[@aria-label='Close' or normalize-space(.)='Close']"))
}
