package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
)

const MongoHome = "/opt/mongodb"

func mongo() *Service {
	if !platform.Current.Proot {
		return containerService(mongoContainer)
	}
	data := filepath.Join(config.DataDir, "mongo")
	bin := MongoHome + "/bin/mongod"
	s := &Service{Key: "mongo", Name: "MongoDB", Aliases: []string{"mongodb", "mongod"},
		LogFile: filepath.Join(config.LogDir, "mongo.log"),
		URL:     func() string { return "mongodb://127.0.0.1:27017" }}
	s.Installed = func() bool { return sys.Exists(bin) }
	s.Running = func() bool { return sys.Pgrep("bin/mongo[d] ") }
	s.start = func() error {
		os.MkdirAll(data, 0o755)
		os.MkdirAll(config.LogDir, 0o777)
		_, err := sys.Output(bin, "--dbpath", data, "--bind_ip", "127.0.0.1", "--port", "27017",
			"--fork", "--logpath", s.LogFile, "--logappend")
		return err
	}
	s.stop = func() error { _, err := sys.Output(bin, "--shutdown", "--dbpath", data); return err }
	s.install = installMongo
	return s
}

func installMongo() error {
	ui.Infof("Procurando a última versão 7.0.x...")
	resp, err := http.Get("https://downloads.mongodb.org/full.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var meta struct {
		Versions []struct {
			Version    string `json:"version"`
			Production bool   `json:"production_release"`
			Downloads  []struct {
				Target, Arch, Edition string
				Archive               struct{ URL, Sha256 string } `json:"archive"`
			} `json:"downloads"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return err
	}
	var url, sum, ver string
	for _, v := range meta.Versions {
		if !v.Production || !strings.HasPrefix(v.Version, "7.0.") {
			continue
		}
		for _, d := range v.Downloads {
			if d.Target == "ubuntu2204" && d.Arch == "aarch64" && d.Edition == "targeted" {
				url, sum, ver = d.Archive.URL, d.Archive.Sha256, v.Version
				break
			}
		}
		if url != "" {
			break
		}
	}
	if url == "" {
		return i18n.Errorf("não achei build 7.0 aarch64")
	}
	ui.Infof("Baixando MongoDB %s...", ver)
	tmp, err := os.CreateTemp("", "mongodb-*.tgz")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), r.Body); err != nil {
		return err
	}
	tmp.Close()
	if got := hex.EncodeToString(h.Sum(nil)); got != sum {
		return i18n.Errorf("checksum não confere (esperado %s, obtido %s)", sum, got)
	}
	ui.Okf("Checksum SHA-256 conferido")
	os.RemoveAll(MongoHome)
	os.MkdirAll(MongoHome, 0o755)
	if out, err := sys.Output("tar", "-xzf", tmp.Name(), "-C", MongoHome, "--strip-components=1"); err != nil {
		return i18n.Errorf("tar: %v %s", err, out)
	}
	if !sys.Have("mongosh") {
		ui.Infof("Instalando mongosh (npm)...")
		if err := sys.Run("", nil, "npm", "install", "-g", "mongosh"); err != nil {
			ui.Warnf("mongosh não instalado: %v", err)
		}
	}
	return nil
}
