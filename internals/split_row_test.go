package internals

import (
	"reflect"
	"testing"
)

// SplitRow cuts a row into one cell per column, at the columns' start offsets.
func TestSplitRow(t *testing.T) {
	pods := []Column{{"NAME", 0}, {"READY", 10}, {"STATUS", 18}, {"RESTARTS", 28}, {"AGE", 39}}
	restarts := []Column{{"NAME", 0}, {"STATUS", 9}, {"RESTARTS", 19}, {"NODE", 33}}
	nameAge := []Column{{"NAME", 0}, {"AGE", 7}}

	tests := []struct {
		name string
		row  string
		cols []Column
		want []string
	}{
		{
			name: "kubectl get pods row",
			row:  "busybox   1/1     Running   0          4m43s",
			cols: pods,
			want: []string{"busybox", "1/1", "Running", "0", "4m43s"},
		},
		{
			name: "spaces inside a cell stay in the cell",
			row:  "api-1    Running   5 (32s ago)   worker-2",
			cols: restarts,
			want: []string{"api-1", "Running", "5 (32s ago)", "worker-2"},
		},
		{
			name: "row shorter than the header gives an empty cell",
			row:  "a-0",
			cols: nameAge,
			want: []string{"a-0", ""},
		},
		{
			name: "row ending exactly where a column starts gives an empty cell",
			row:  "a-0    ",
			cols: nameAge,
			want: []string{"a-0", ""},
		},
		{
			name: "empty row gives all empty cells",
			row:  "",
			cols: pods,
			want: []string{"", "", "", "", ""},
		},
		{
			name: "single column returns the trimmed row",
			row:  "  hello  ",
			cols: []Column{{"NAME", 0}},
			want: []string{"hello"},
		},
		{
			name: "last column keeps inner spaces and trims the edges",
			row:  "a-0    some  text here  ",
			cols: nameAge,
			want: []string{"a-0", "some  text here"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SplitRow(tc.row, tc.cols)
			if err != nil {
				t.Fatalf("SplitRow(%q) returned an error: %v", tc.row, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SplitRow(%q)\n got: %q\nwant: %q", tc.row, got, tc.want)
			}
		})
	}
}

// A negative Start can never be a real column position, so SplitRow must
// return an error instead of slicing with it.
func TestSplitRowRejectsNegativeStart(t *testing.T) {
	got, err := SplitRow("abc", []Column{{"A", -1}, {"B", 4}})
	if err == nil {
		t.Fatalf("SplitRow returned %q, want an error", got)
	}
}

// A cell ends where the next column starts, so Start offsets must not go
// backwards: row[10:3] would panic. SplitRow must return an error instead.
func TestSplitRowRejectsDescendingStarts(t *testing.T) {
	got, err := SplitRow("0123456789abcdefghij", []Column{{"A", 10}, {"B", 3}})
	if err == nil {
		t.Fatalf("SplitRow returned %q, want an error", got)
	}
}
