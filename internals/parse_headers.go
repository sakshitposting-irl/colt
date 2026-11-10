package internals

import "errors"

// Column is one header word and the byte offset where it starts in the header
// line. A column ends where the next one starts; the last one runs to the end
// of the line.

// example: "  Name  Age  City" has columns:
//
//	Name: 2
//	Age: 8
//	City: 13
type Column struct {
	Name  string
	Start int
}

// ParseHeader finds the columns of a table from its header line: each run of
// non-space characters is a column, starting at its first character.
// The name keeps the header's original spelling.
func ParseHeader(line string) ([]Column, error) {
	var cols []Column
	start := -1 // start of the word we're inside, or -1 between words
	for i := 0; i <= len(line); i++ {
		// a word is inword if we're not at the end of the line and the current character is not a space
		inWord := i < len(line) && line[i] != ' '
		// If we're inside a word and haven't recorded its start, record it.
		if inWord && start < 0 {
			start = i
		}
		if !inWord && start >= 0 {
			// We've reached the end of a word, so record it.
			// storing a {k,v} pair of the word and its starting index
			// storing this so that we can later extract the column data from the table rows using these indices.
			cols = append(cols, Column{Name: line[start:i], Start: start})
			start = -1
		}
	}
	if len(cols) == 0 {
		return nil, errors.New("header line is empty")
	}
	return cols, nil
}
