package services

import (
	"net"
	"strings"
	"testing"
)

func fakeServer(t *testing.T, reply []byte) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, 8)
			conn.Read(buf)
			if reply != nil {
				conn.Write(reply)
			}
			conn.Close()
		}
	}()
	return ln.Addr().String()[strings.LastIndex(ln.Addr().String(), ":")+1:]
}

func TestReadyOnlyWhenProtocolResponds(t *testing.T) {
	if !postgresReady(fakeServer(t, []byte("N"))) {
		t.Fatal("PostgreSQL answered N and should be ready")
	}
	if !amqpReady(fakeServer(t, []byte{1, 0, 0})) {
		t.Fatal("RabbitMQ answered connection.start and should be ready")
	}
	if postgresReady(fakeServer(t, nil)) || amqpReady(fakeServer(t, nil)) {
		t.Fatal("a port that accepts and closes without answering, like docker-proxy's, is not ready")
	}
}
