// goadifsplit: split an ADIF file into multiple files
// by Kenji Rikitake, JJ1BDX
// Usage: goadifsplit [-f infile] [-l lines] [-a suffix_length]
//                    [-p prefix] [-e extension]
//
// Coding convention:
// Use reader for reading each record (with ADIFReader)
// Use writer for writing each record (with ADIFWriter)

package main

import (
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"io"
	"os"
	"strconv"
	"strings"
)

func numtosuffix(n uint, suflen uint) string {
	// Use decimal digits only
	nstr := strconv.FormatUint(uint64(n), 10)
	nlen := len(nstr)
	slen := int(suflen)
	if nlen > slen {
		fmt.Fprintln(os.Stderr, "Error: numtosuffix length overflow")
		os.Exit(1)
	}
	if nlen < slen {
		addlen := slen - nlen
		outstr := strings.Repeat("0", addlen) + nstr
		return outstr
	} else {
		return nstr
	}
}

func genfilename(n uint, suflen uint, prefix string, extension string) string {
	return prefix + numtosuffix(n, suflen) + "." + extension
}

func main() {
	var infile = flag.String("f", "", "input file (stdout in none)")
	var filelines = flag.Uint("l", 100, "lines per output file (default 100)")
	var indexlength = flag.Uint("a", 4, "suffix length (default 4)")
	var outprefix = flag.String("p", "x", "output file prefix (default 'x')")
	var extension = flag.String("e", "adi", "output file extension without period (default 'adi'")

	var fp *os.File
	var err error

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintf(flag.CommandLine.Output(), "goadifsplit: split an ADIF file into multiple files\n\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-f infile] [-l lines] [-a length] [-p prefix] [-e extension]\n", execname)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *filelines == 0 {
		fmt.Fprintln(os.Stderr, "Error: lines per output file must be a positive number")
		return
	}

	if *infile == "" {
		fp = os.Stdin
	} else {
		fp, err = os.Open(*infile)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	}

	reader := adifparser.NewADIFReader(fp)

	var outfilenum uint
	outfilenum = 0

	var writer adifparser.ADIFWriter
	var writefp *os.File

	endoffile := false

	for !endoffile {
		var outfilename = genfilename(outfilenum, *indexlength, *outprefix, *extension)
		if _, err := os.Stat(outfilename); os.IsNotExist(err) {
			// File does not exist: create it
			writefp, err = os.Create(outfilename)
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				return
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: file %s already exists\n", outfilename)
			return
		}
		writer = adifparser.NewADIFWriter(writefp)

		comment := fmt.Sprintf("goadifsplit file %d\n", outfilenum)
		if writer.SetComment(comment) != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		maxlines := int(*filelines)
		linecount := 0

		for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
			if err != nil {
				if err != io.EOF {
					fmt.Fprint(os.Stderr, err)
				}
				endoffile = true
				break // when io.EOF break the loop!
			}
			// Output the record
			writer.WriteRecord(record)
			linecount = linecount + 1
			if linecount >= maxlines {
				break // break the loop
			}
		}

		// Flush and close the output
		writer.Flush()
		if writefp != os.Stdout {
			writefp.Close()
		}
		outfilenum = outfilenum + 1
	}

	fmt.Fprintf(os.Stderr, "Total records: %d\n", reader.RecordCount())
	fmt.Fprintf(os.Stderr, "Total number of files: %d\n", outfilenum)
}
