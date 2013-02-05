// loop through ERF records wiht shifting ERF record timestamp
// in the first pass collects stats and calculates offset for the following passes

package main

import (
	"fmt"
	"os"
    "flag"
    "endace/erf/multiplier"
)

var (
    // offset in sec
    offset int
)

func usage() {
    fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()

    fmt.Fprintf(os.Stderr, `
Example:

  %s -off -30000000 < ~/small.erf  | nc 192.168.132.109 9999

  It will shift timestamps back into the past, loop through all ERF
  records in the file 'small.erf' and "netcats" them to the host port 9999
`, os.Args[0])    
}

func init() {
    flag.Usage = usage

    flag.IntVar( &offset, "off", 0, "Initial offset in sec")
    flag.Parse()
}

func assert(t bool, s string) {
    if !t {
        panic(s)
    }
}

func  main() {

    multiplier.Loop( os.Stdin, os.Stdout, offset)

}
