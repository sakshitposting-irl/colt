package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const usage = `Usage:
  colt [-s columns] [-w condition]... [--no-headers] [file...]
  colt --explain [file...]
  colt help | version | learn

Flags:
`

// multiFlag is a flag that can be given more than once.
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			fmt.Print(usage)
			flagSet(new(string), new(multiFlag), new(bool), new(bool)).PrintDefaults()
			return
		case "version":
			fmt.Println("Colt version 1.0.0")
			return
		case "learn":
			fmt.Fprintln(os.Stderr, "colt: learn is not implemented yet")
			os.Exit(1)
		}
	}

	var (
		sel       string
		where     multiFlag
		noHeaders bool
		explain   bool
	)
	fs := flagSet(&sel, &where, &noHeaders, &explain)
	fs.Parse(os.Args[1:])

	if err := numberLines(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "colt: %v\n", err)
		os.Exit(1)
	}
}

// numberLines writes each line of in to out as "<n>: <line>", n starting at 1.
func numberLines(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	n := 0
	for scanner.Scan() {
		n++
		if _, err := fmt.Fprintf(out, "%d: %s\n", n, scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func flagSet(sel *string, where *multiFlag, noHeaders, explain *bool) *flag.FlagSet {
	fs := flag.NewFlagSet("colt", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		fs.PrintDefaults()
	}
	fs.StringVar(sel, "s", "", "columns to print, in this order (default: all)")
	fs.StringVar(sel, "select", "", "same as -s")
	fs.Var(where, "w", "keep only rows matching the condition; repeat to AND")
	fs.Var(where, "where", "same as -w")
	fs.BoolVar(noHeaders, "no-headers", false, "don't print the header row")
	fs.BoolVar(explain, "explain", false, "show how colt split the table instead of printing it")
	return fs
}
