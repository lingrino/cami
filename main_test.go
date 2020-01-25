package main

import (
	"fmt"
	"testing"
)

// TestMainInput is the input for our table testing
type TestMainInput struct {
	executeF func(string) error
}

// TestMain tests the main() function
func TestMain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		give func(string) error
	}{
		{
			name: "valid",
			give: func(string) error { return nil },
		},
		{
			name: "error",
			give: func(string) error { return fmt.Errorf("FAIL") },
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			executeF = tt.give
			main()
		})
	}
}
