package internals

import (
	"fmt"
	"slices"
	"strings"
)

// SplitRow cuts a data row into one cell per column, slicing at the columns'
// start offsets: cell i runs from cols[i].Start to cols[i+1].Start, and the
// last cell runs to the end of the row. Spaces around each cell are trimmed.
//
// On success it returns exactly len(cols) cells. A row too short to reach a
// column gives "" for that cell instead of panicking. Equal neighbouring
// offsets are allowed and give an empty cell.
//
// It returns an error, and no cells, if any Start is negative or if the Start
// offsets are not in ascending order, since either would make a slice bound
// invalid.
//
// SplitRow does not check that the row fits the columns: text that crosses a
// boundary is sliced silently.
func SplitRow(row string, cols []Column) ([]string, error) {

	colList := make([]int, len(cols))
	for i, c := range cols {
		colList[i] = c.Start
		if colList[i] < 0 {
			return nil, fmt.Errorf("column %d has negative start offset", i)
		}
	}
	if !slices.IsSorted(colList) {
		return nil, fmt.Errorf("columns are not sorted by start offset")
	}

	cells := make([]string, len(cols))
	for i, c := range cols {
		start := min(c.Start, len(row))
		end := len(row)
		if i+1 < len(cols) {
			end = min(cols[i+1].Start, len(row))
		}
		cells[i] = strings.Trim(row[start:end], " ")
	}
	return cells, nil
}
