// goadifsplit: split an ADIF file into multiple files
// by Kenji Rikitake, JJ1BDX
// Usage: goadifsplit [-f infile] [-l lines] [-a suffix_length]
//                    [-p prefix] [-e extension]
//
// -p and -e are concatenated literally into the output path:
//   <prefix><zero-padded-index>.<extension>
// They are written by the invoking user, so a prefix containing path
// separators ("foo/", "../bar/") writes the split files into those
// directories. Treat -p / -e as paths under your control.
//
// Coding convention:
// Use reader for reading each record (with ADIFReader)
// Use writer for writing each record (with ADIFWriter)

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"io"
	"os"
	"strconv"
	"strings"
)

func numtosuffix(n uint, suflen uint) (string, error) {
	// Use decimal digits only
	nstr := strconv.FormatUint(uint64(n), 10)
	nlen := len(nstr)
	slen := int(suflen)
	if nlen > slen {
		return "", errors.New("numtosuffix length overflow")
	}
	if nlen < slen {
		addlen := slen - nlen
		return strings.Repeat("0", addlen) + nstr, nil
	}
	return nstr, nil
}

func genfilename(n uint, suflen uint, prefix string, extension string) (string, error) {
	suffix, err := numtosuffix(n, suflen)
	if err != nil {
		return "", err
	}
	return prefix + suffix + "." + extension, nil
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
		os.Exit(2)
	}

	if *infile == "" {
		fp = os.Stdin
	} else {
		fp, err = os.Open(*infile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer fp.Close()
	}

	reader := adifparser.NewADIFReader(fp)

	var outfilenum uint = 0

	// Peek at the first record so we can avoid creating an empty
	// trailing file when the input ends on an exact filelines boundary.
	pendingRecord, pendingErr := reader.ReadRecord()
	for pendingRecord != nil {
		outfilename, gerr := genfilename(outfilenum, *indexlength, *outprefix, *extension)
		if gerr != nil {
			fmt.Fprintln(os.Stderr, gerr)
			os.Exit(1)
		}
		// O_EXCL: atomic create-if-absent that refuses to follow a
		// final-component symlink. Replaces the per-file Stat/Create
		// TOCTOU race, which is wider here because of the loop.
		writefp, oerr := os.OpenFile(outfilename,
			os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if oerr != nil {
			fmt.Fprintln(os.Stderr, oerr)
			os.Exit(1)
		}
		writer := adifparser.NewADIFWriter(writefp)

		comment := fmt.Sprintf("goadifsplit file %d\n", outfilenum)
		if cerr := writer.SetComment(comment); cerr != nil {
			fmt.Fprintln(os.Stderr, cerr)
			writefp.Close()
			os.Exit(1)
		}

		maxlines := int(*filelines)
		linecount := 0

		// Drain the pending record from the previous outer iteration,
		// then keep reading until we hit maxlines or EOF.
		for pendingRecord != nil && linecount < maxlines {
			if werr := writer.WriteRecord(pendingRecord); werr != nil {
				fmt.Fprintln(os.Stderr, werr)
				writefp.Close()
				os.Exit(1)
			}
			linecount++
			pendingRecord, pendingErr = reader.ReadRecord()
		}

		if ferr := writer.Flush(); ferr != nil {
			fmt.Fprintln(os.Stderr, ferr)
			writefp.Close()
			os.Exit(1)
		}
		if cerr := writefp.Close(); cerr != nil {
			fmt.Fprintln(os.Stderr, cerr)
			os.Exit(1)
		}
		outfilenum++
	}

	if pendingErr != nil && pendingErr != io.EOF {
		fmt.Fprintln(os.Stderr, pendingErr)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Total records: %d\n", reader.RecordCount())
	fmt.Fprintf(os.Stderr, "Total number of files: %d\n", outfilenum)
}
