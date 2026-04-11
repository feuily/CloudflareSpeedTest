package task

import "testing"

func TestIsValidHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{name: "zero uses default", statusCode: 0, want: false},
		{name: "below range", statusCode: 99, want: false},
		{name: "lower bound", statusCode: 100, want: true},
		{name: "success", statusCode: 200, want: true},
		{name: "upper bound", statusCode: 599, want: true},
		{name: "above range", statusCode: 600, want: false},
	}

	for _, test := range tests {
		if got := isValidHTTPStatusCode(test.statusCode); got != test.want {
			t.Fatalf("%s: expected %v, got %v", test.name, test.want, got)
		}
	}
}