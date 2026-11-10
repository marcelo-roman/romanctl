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

func TestProntoSoQuandoOProtocoloResponde(t *testing.T) {
	if !postgresReady(fakeServer(t, []byte("N"))) {
		t.Fatal("PostgreSQL respondeu N e deveria estar pronto")
	}
	if !amqpReady(fakeServer(t, []byte{1, 0, 0})) {
		t.Fatal("RabbitMQ respondeu connection.start e deveria estar pronto")
	}
	if postgresReady(fakeServer(t, nil)) || amqpReady(fakeServer(t, nil)) {
		t.Fatal("uma porta que aceita e fecha sem responder, como a do docker-proxy, não está pronta")
	}
}
