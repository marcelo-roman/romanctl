package services

import (
	"net"
	"time"
)

// O docker-proxy abre a porta do host antes de o serviço dentro do container escutar: abrir a
// conexão não prova nada. Estes testes conversam o começo do protocolo e só dão o serviço por
// pronto quando ele responde.

type readiness func(port string) bool

// postgresReady manda o pedido de SSL do protocolo do PostgreSQL; um servidor no ar responde S ou N.
func postgresReady(port string) bool {
	return answers(port, []byte{0, 0, 0, 8, 0x04, 0xd2, 0x16, 0x2f}, func(b []byte) bool { return b[0] == 'S' || b[0] == 'N' })
}

// amqpReady manda o cabeçalho do AMQP 0-9-1; um RabbitMQ no ar responde com o frame connection.start.
func amqpReady(port string) bool {
	return answers(port, []byte("AMQP\x00\x00\x09\x01"), func(b []byte) bool { return b[0] == 1 })
}

func answers(port string, hello []byte, ok func([]byte) bool) bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write(hello); err != nil {
		return false
	}
	b := make([]byte, 1)
	if n, err := conn.Read(b); err != nil || n == 0 {
		return false
	}
	return ok(b)
}
