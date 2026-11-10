package webui

import (
	"io/fs"
	"net/http"
	"strings"
	"testing"
)

func TestADistEmbutidaTemOIndexEOsAssets(t *testing.T) {
	index, err := fs.ReadFile(dist, "dist/index.html")
	if err != nil {
		t.Fatalf("dist/index.html ausente: rode npm run build em web/ (%v)", err)
	}
	if !strings.Contains(string(index), "</head>") {
		t.Fatal("dist/index.html sem </head>: o token não teria onde entrar")
	}
	assets, err := fs.ReadDir(dist, "dist/assets")
	if err != nil || len(assets) == 0 {
		t.Fatalf("dist/assets vazio: rode npm run build em web/ (%v)", err)
	}
	_, h := testServer(t, "/bin/true")
	for _, a := range assets {
		rec := call(h, "GET", "/assets/"+a.Name(), "localhost:7777", "", "")
		if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
			t.Fatalf("/assets/%s: %d %q", a.Name(), rec.Code, rec.Header().Get("Cache-Control"))
		}
	}
}
