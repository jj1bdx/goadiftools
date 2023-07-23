// goadifgrep: search specified ADIF field with a regex and output matched ADIF record
// by Kenji Rikitake, JJ1BDX
// Usage: goadifgrep [-v] [-f infile] [-o outfile] field regex
// Note: field name is case insensitive
// Note 2: regex is Go RE2 as defined in Go regexp package
//         Use "(?i)" flag prefix for case-insensitive matching

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/jj1bdx/adifparser"
)

func main() {
	var infile = flag.String("f", "", "input file (stdout in none)")
	var outfile = flag.String("o", "", "output file (stdout if none)")
	var invertmatch = flag.Bool("v", false, "invert match if specified")

	var fp *os.File
	var err error

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintln(flag.CommandLine.Output(),
			"goadifgrep: search specified ADIF field with a regex and output matched ADIF record")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-v] [-f infile] [-o outfile] field regex\n", execname)
		fmt.Fprintf(flag.CommandLine.Output(),
			"Note: field name is case insensitive\n"+
				"Note 2: regex is Go RE2 as defined in Go regexp package\n"+
				"        Use \"(?i)\" flag prefix for case-insensitive matching\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *infile == "" {
		fp = os.Stdin
	} else {
		fp, err = os.Open(*infile)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	}

	var writer adifparser.ADIFWriter
	var writefp *os.File
	if *outfile != "" {
		writefp, err = os.Create(*outfile)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		writer = adifparser.NewADIFWriter(writefp)
	} else {
		writefp = nil
		writer = adifparser.NewADIFWriter(os.Stdout)
	}

	if writer.SetComment("goadifgrep\n") != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	cliargs := flag.Args()
	if len(cliargs) != 2 {
		fmt.Fprint(os.Stderr, "Error: incorrect arguments\n")
		return
	}
	var fieldname = strings.ToLower(cliargs[0])
	var regpattern = regexp.MustCompile(cliargs[1])

	reader := adifparser.NewADIFReader(fp)
	for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
		if err != nil {
			if err != io.EOF {
				fmt.Fprint(os.Stderr, err)
			}
			break // when io.EOF break the loop!
		}

		// obtain selected field value
		fieldvalue, err := record.GetValue(fieldname)
		if err == adifparser.ErrNoSuchField {
			fieldvalue = ""
		} else if err != nil {
			fmt.Fprint(os.Stderr, err)
			break
		}

		// Check regex pattern matching
		matched := regpattern.MatchString(fieldvalue)
		var selected bool
		if *invertmatch {
			selected = !matched
		} else {
			selected = matched
		}

		// Output selected record
		if selected {
			writer.WriteRecord(record)
		}
	}

	// Flush and close the output
	writer.Flush()
	if writefp != os.Stdout {
		writefp.Close()
	}
}
