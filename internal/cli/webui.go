package cli

import (
	"strconv"

	"romanctl/internal/ui"
	"romanctl/internal/webui"
)

func webUICommand(flags map[string]bool) error {
	port := webui.DefaultPort
	if v := flagVals["port"]; v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 65535 {
			return ui.Usagef("porta inválida: %s", v)
		}
		port = n
	}
	return webui.Serve(webui.Options{Port: port, Open: flags["open"], Version: version})
}
