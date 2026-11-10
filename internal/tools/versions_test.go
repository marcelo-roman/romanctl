package tools

import (
	"slices"
	"testing"
)

func TestVersoesEmOrdemNumerica(t *testing.T) {
	got := plain([]string{"v1.9.3", "v1.10.0", "v1.10.0-rc.1", "v2.0.0-beta", "v1.2.0", "nightly", "v1.9.10"})
	var labels []string
	for _, v := range got {
		labels = append(labels, v.Label)
	}
	if want := []string{"1.10.0", "1.9.10", "1.9.3", "1.2.0"}; !slices.Equal(labels, want) {
		t.Fatalf("%v", labels)
	}
	if got[0].ID != "v1.10.0" {
		t.Fatal("o ID guarda a tag como ela é, com o v")
	}
}

func TestPreReleaseFicaDeFora(t *testing.T) {
	for _, v := range []string{"1.0.0-rc1", "2.0.0b1", "3.1.0a2", "1.2.0.dev3", "1.5.0+ent", "27.0.0-alpha.0", "4.0.0-next.1"} {
		if stable(v) {
			t.Errorf("%s deveria ser pré-release", v)
		}
	}
	for _, v := range []string{"1.0.0", "v0.51.0", "1.7.1-3ubuntu0.24.04.2", "kustomize/v5.8.3"} {
		if !stable(v) {
			t.Errorf("%s é estável", v)
		}
	}
}

func TestTrocarDeVersaoEhReinstalar(t *testing.T) {
	if UpToDate("helm v3.16.2+g1a2b", "3.15.0") {
		t.Fatal("versões diferentes não estão em dia")
	}
	if !UpToDate("v0.51.0", "0.51.0") {
		t.Fatal("o v não muda a versão")
	}
}
