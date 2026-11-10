package shared

import (
	"crypto/rand"
	"fmt"
	"regexp"
)

type ID string

var idFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func NewID() ID {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return ID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]))
}

func ParseID(s string) (ID, error) {
	if !idFormat.MatchString(s) {
		return "", Rule("Identificador inválido.")
	}
	return ID(s), nil
}
