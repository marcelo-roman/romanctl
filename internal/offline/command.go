package offline

import (
	"romanctl/internal/i18n"
)

func Command(sub string, flags map[string]bool) error {
	switch sub {
	case "", "status":
		offlineStatus()
		return nil
	case "prep", "preparar":
		return offlinePrep(flags["force"])
	case "test", "testar":
		return offlineTest()
	}
	return i18n.Errorf("use: romanctl offline [status|prep|test]")
}
