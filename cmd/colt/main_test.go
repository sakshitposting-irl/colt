package main

import (
	"bytes"
	"strings"
	"testing"
)

// Rung 1: read input and print each line with its number.
//
// Write this function in main.go:
//
//	func numberLines(in io.Reader, out io.Writer) error
//
// Each output line is "<n>: <original line>", with n starting at 1.
func TestNumberLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "two lines",
			input: "alpha\nbeta\n",
			want:  "1: alpha\n2: beta\n",
		},
		{
			name:  "empty input prints nothing",
			input: "",
			want:  "",
		},
		{
			name:  "last line has no trailing newline",
			input: "alpha\nbeta",
			want:  "1: alpha\n2: beta\n",
		},
		{
			name:  "blank lines still get a number",
			input: "alpha\n\nbeta\n",
			want:  "1: alpha\n2: \n3: beta\n",
		},
		{
			name:  "spaces inside a line are kept",
			input: "    inet 127.0.0.1/8 scope host lo\n",
			want:  "1:     inet 127.0.0.1/8 scope host lo\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := numberLines(strings.NewReader(tc.input), &out)
			if err != nil {
				t.Fatalf("numberLines returned an error: %v", err)
			}
			if got := out.String(); got != tc.want {
				t.Errorf("numberLines(%q)\n got: %q\nwant: %q", tc.input, got, tc.want)
			}
		})
	}
}
