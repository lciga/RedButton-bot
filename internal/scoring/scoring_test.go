package scoring

import "testing"

func TestCalculate(t *testing.T) {
	tests := []struct {
		name       string
		solveCount int64
		want       int
	}{
		{name: "without solves", solveCount: 0, want: 100},
		{name: "first solve", solveCount: 1, want: 97},
		{name: "decay reached", solveCount: 5, want: 10},
		{name: "minimum points", solveCount: 200, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Calculate(100, 10, 5, tt.solveCount); got != tt.want {
				t.Errorf("Calculate() = %d, want %d", got, tt.want)
			}
		})
	}
}
