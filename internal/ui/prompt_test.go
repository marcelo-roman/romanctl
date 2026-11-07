package ui

import (
	"errors"
	"slices"
	"testing"
)

// run simula respostas: cada passo devolve o próximo erro da sua fila.
func run(answers map[int][]error) ([]int, error) {
	var visited []int
	step := func(i int) func() error {
		return func() error {
			visited = append(visited, i)
			q := answers[i]
			if len(q) == 0 {
				return nil
			}
			answers[i] = q[1:]
			return q[0]
		}
	}
	err := Wizard(step(0), step(1), step(2))
	return visited, err
}

func TestWizardAvanca(t *testing.T) {
	visited, err := run(map[int][]error{})
	if err != nil || !slices.Equal(visited, []int{0, 1, 2}) {
		t.Fatalf("%v %v", visited, err)
	}
}

func TestWizardVoltaUmPasso(t *testing.T) {
	visited, err := run(map[int][]error{2: {errBack}})
	if err != nil || !slices.Equal(visited, []int{0, 1, 2, 1, 2}) {
		t.Fatalf("%v %v", visited, err)
	}
}

func TestWizardNoPrimeiroPassoCancela(t *testing.T) {
	if _, err := run(map[int][]error{0: {errBack}}); !errors.Is(err, errBack) {
		t.Fatalf("%v", err)
	}
}

// Sem a direção, voltar para um passo pulado devolveria a pessoa ao passo de onde saiu.
func TestWizardPassoPuladoNaoPrende(t *testing.T) {
	visited, err := run(map[int][]error{1: {ErrSkip, ErrSkip}, 2: {errBack}})
	if err != nil || !slices.Equal(visited, []int{0, 1, 2, 1, 0, 1, 2}) {
		t.Fatalf("%v %v", visited, err)
	}
	if _, err := run(map[int][]error{0: {ErrSkip, ErrSkip}, 1: {errBack}}); !errors.Is(err, errBack) {
		t.Fatalf("voltar até um primeiro passo pulado deveria cancelar: %v", err)
	}
}

func TestWizardOutroErroPara(t *testing.T) {
	boom := errors.New("x")
	visited, err := run(map[int][]error{1: {boom}})
	if !errors.Is(err, boom) || !slices.Equal(visited, []int{0, 1}) {
		t.Fatalf("%v %v", visited, err)
	}
}
