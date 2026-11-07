package sys

import (
	"net"
	"net/http"
	"os"
	"time"
)

func IsOnline() bool {
	if os.Getenv("ROMANCTL_OFFLINE") != "" {
		return false
	}
	c, err := net.DialTimeout("tcp", "registry.npmjs.org:443", 2*time.Second)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func HttpUp(url string) bool {
	c := http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := c.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}
