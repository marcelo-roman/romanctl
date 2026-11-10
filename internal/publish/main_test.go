package publish

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	if dir, err := filepath.Abs("../../templates"); err == nil {
		os.Setenv("ROMANCTL_TEMPLATES", dir)
	}
	os.Exit(m.Run())
}
