package shared

import "errors"

var ErrNotFound = errors.New("não encontrado")

type RuleError struct {
	message string
}

func (e *RuleError) Error() string { return e.message }

func Rule(message string) error { return &RuleError{message: message} }
