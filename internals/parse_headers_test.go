package internals

import (
	"reflect"
	"testing"
)

// ParseHeader finds where each column starts from the header line.
func TestParseHeader(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    []Column
		wantErr bool
	}{
		{
			name: "kubectl get pods header",
			line: "NAME      READY   STATUS    RESTARTS   AGE",
			want: []Column{
				{"NAME", 0}, {"READY", 10}, {"STATUS", 18}, {"RESTARTS", 28}, {"AGE", 39},
			},
		},
		{
			name: "single column",
			line: "NAME",
			want: []Column{{"NAME", 0}},
		},
		{
			name: "leading spaces shift the start",
			line: "  NAME   AGE",
			want: []Column{{"NAME", 2}, {"AGE", 9}},
		},
		{
			name: "trailing spaces are ignored",
			line: "NAME   AGE   ",
			want: []Column{{"NAME", 0}, {"AGE", 7}},
		},
		{
			name: "name keeps original case",
			line: "Name  sTaTus",
			want: []Column{{"Name", 0}, {"sTaTus", 6}},
		},
		{name: "empty line is an error", line: "", wantErr: true},
		{name: "spaces only is an error", line: "    ", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseHeader(tc.line)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseHeader(%q) = %v, want an error", tc.line, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseHeader(%q) returned an error: %v", tc.line, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseHeader(%q)\n got: %v\nwant: %v", tc.line, got, tc.want)
			}
		})
	}
}
