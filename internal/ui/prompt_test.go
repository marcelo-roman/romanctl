package ui

import (
	"errors"
	"slices"
	"testing"
)

// run simulates answers: each step returns the next error from its queue.
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

func TestWizardAdvances(t *testing.T) {
	visited, err := run(map[int][]error{})
	if err != nil || !slices.Equal(visited, []int{0, 1, 2}) {
		t.Fatalf("%v %v", visited, err)
	}
}

func TestWizardGoesBackOneStep(t *testing.T) {
	visited, err := run(map[int][]error{2: {errBack}})
	if err != nil || !slices.Equal(visited, []int{0, 1, 2, 1, 2}) {
		t.Fatalf("%v %v", visited, err)
	}
}

func TestWizardBackOnFirstStepCancels(t *testing.T) {
	if _, err := run(map[int][]error{0: {errBack}}); !errors.Is(err, errBack) {
		t.Fatalf("%v", err)
	}
}

// Without the direction, going back to a skipped step would return the user to the step they left.
func TestWizardSkippedStepDoesNotTrap(t *testing.T) {
	visited, err := run(map[int][]error{1: {ErrSkip, ErrSkip}, 2: {errBack}})
	if err != nil || !slices.Equal(visited, []int{0, 1, 2, 1, 0, 1, 2}) {
		t.Fatalf("%v %v", visited, err)
	}
	if _, err := run(map[int][]error{0: {ErrSkip, ErrSkip}, 1: {errBack}}); !errors.Is(err, errBack) {
		t.Fatalf("going back to a skipped first step should cancel: %v", err)
	}
}

func TestWizardOtherErrorStops(t *testing.T) {
	boom := errors.New("x")
	visited, err := run(map[int][]error{1: {boom}})
	if !errors.Is(err, boom) || !slices.Equal(visited, []int{0, 1}) {
		t.Fatalf("%v %v", visited, err)
	}
}
