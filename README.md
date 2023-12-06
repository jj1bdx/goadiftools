# goadiftools: ADIF Tools in Go

Go tools for ADIF ADI files

## Tools

* goadifcab: output Cabrillo QSO log entries for given ADIF records
* goadifcsv: output specified ADIF fields from the input ADIF records in CSV format
* goadifdelf: delete specified ADIF fields from the input ADIF records
* goadifdedupe: dump QSOs WITH deduping (eliminating dupe QSOs)
* goadifdump: skeleton for further writing the code
* goadifdxcc: add missing DXCC fields using godxcc
* goadifdxcccl: add missing DXCC fields using gocldb
* goadifgrep: search specified ADIF field with a regex and output matched ADIF record
* goadifstat: obtain QSO statistics
* goadiftime: sort and filter QSOs by QSO\_DATE/TIME\_ON fields
* noasciitostar: convert non-ASCII UTF-8 letters to "\*" of the same byte length
  - This text filter guarantees the result only contains ASCII letters

## Things to do before compilation

```shell
go mod init github.com/jj1bdx/goadiftools
go mod tidy
```

## How to compile

Do `./buildall.sh`

## Required libraries

* https://github.com/jj1bdx/adifparser
* https://github.com/jj1bdx/gocldb
* https://github.com/jj1bdx/godxcc

## Usage examples

See the contents in `examples/` for the usage examples.

## License

BSD 2-clause License
