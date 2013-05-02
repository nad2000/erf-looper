// loop through ERF records wiht shifting ERF record timestamp
// in the first pass collects stats and calculates offset for the following passes

package main

import (
	"endace/erf/multiplier"
	"flag"
	"fmt"
	"os"
)

var (
	// offset in sec
	offset int
	// dry run - don't write into output and completing after the first pass
	dry_run bool
	// generated traffic limit in seconds
	limit int
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, `
Examples:

  %s -off -30000000 < ~/small.erf  | nc 192.168.132.109 9999

  It will shift timestamps back into the past, loop through all ERF
  records in the file 'small.erf' and "netcats" them to the host port 9999

  %s -dry -off -30000000 < ~/small.erf

  Dry run without writing to the output, and completing after the first pass

`, os.Args[0], os.Args[0])
}

func init() {
	flag.Usage = usage

	flag.IntVar(
		&offset,
		"off",
		0,
		"Initial offset in sec")
	flag.IntVar(
		&limit,
		"lim",
		0,
		"Limit in sec of generated traffic (default: 0 - no limit)")
	flag.BoolVar(
		&dry_run,
		"dry",
		false,
		"Dry run without writing to the output, and completing after the first pass")
	flag.Parse()
}

func assert(t bool, s string) {
	if !t {
		panic(s)
	}
}

func main() {

	multiplier.Loop(os.Stdin, os.Stdout, offset, dry_run, limit)

}

// vim: foldmethod=syntax
