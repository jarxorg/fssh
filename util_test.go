package fssh

import (
	"testing"
)

func TestIsPattern(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{
			pattern: "abc",
			want:    false,
		}, {
			pattern: "*.txt",
			want:    true,
		}, {
			pattern: "**/*.txt",
			want:    true,
		}, {
			pattern: "?.txt",
			want:    true,
		}, {
			pattern: "[a-z].txt",
			want:    true,
		}, {
			pattern: "[].txt",
			want:    true,
		},
	}
	for i, test := range tests {
		got := IsPattern(test.pattern)
		if got != test.want {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}
