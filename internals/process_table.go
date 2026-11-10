package internals

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ProcessTable reads a table from in and writes it to out: the header line
// first, then each data row as soon as it is read. Rows are rebuilt from their
// cells, so each column keeps the width it had in the input.
//
// It returns an error if the input is empty, if the first line is blank, or if
// reading or writing fails. Rows written before an error stay written.
func ProcessTable(in io.Reader, out io.Writer) error {
	// A Scanner hands us one line at a time with the newline already removed,
	// and it also returns a last line that has no trailing newline. A
	// bufio.Reader's ReadString keeps the newline and returns the last line
	// together with io.EOF, which we would have to handle ourselves.
	scanner := bufio.NewScanner(in)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		return errors.New("no input")
	}

	cols, err := ParseHeader(scanner.Text())
	if err != nil {
		return errors.New("first line is blank, the header must be on line 1")
	}

	names := make([]string, len(cols))
	// write the header line first, so that the output is valid even if reading fails later, len(cols) is guaranteed to be > 0 because ParseHeader returns an error if the header is blank.
	for i, c := range cols {
		names[i] = c.Name
	}
	if _, err := fmt.Fprintln(out, formatRow(names, cols)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	lineNo := 1 // the header was line 1

	// read each line, split it into cells, and write it to the output
	for scanner.Scan() {
		lineNo++
		cells, err := SplitRow(scanner.Text(), cols)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
		if _, err := fmt.Fprintln(out, formatRow(cells, cols)); err != nil {
			return fmt.Errorf("failed to write line %d: %w", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read line %d: %w", lineNo+1, err)
	}
	return nil
}

// formatRow joins cells into one line. Every cell but the last is padded with
// spaces to its column's width, the distance to the next column's Start. The
// last cell is written as is, so an empty last cell leaves the padding of the
// column before it at the end of the line.
func formatRow(cells []string, cols []Column) string {
	var b strings.Builder
	for i, cell := range cells {
		b.WriteString(cell)
		if i+1 < len(cols) {
			width := cols[i+1].Start - cols[i].Start
			b.WriteString(strings.Repeat(" ", max(0, width-len(cell))))
		}
	}
	return b.String()
}
