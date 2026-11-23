package services

import (
	"net"
	"time"
)

// docker-proxy opens the host port before the service inside the container listens: opening the
// connection proves nothing. These checks speak the start of the protocol and only consider the
// service ready once it answers.

type readiness func(port string) bool

// postgresReady sends the PostgreSQL protocol's SSL request; a live server answers S or N.
func postgresReady(port string) bool {
	return answers(port, []byte{0, 0, 0, 8, 0x04, 0xd2, 0x16, 0x2f}, func(b []byte) bool { return b[0] == 'S' || b[0] == 'N' })
}

// amqpReady sends the AMQP 0-9-1 header; a live RabbitMQ answers with the connection.start frame.
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
