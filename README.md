# goadiftools: ADIF Tools in Go

Go tools for ADIF ADI files

## Tools

* goadifcsv: output specified ADIF fields from the input ADIF records in CSV format
* goadifdelf: delete specified ADIF fields from the input ADIF records
* goadifdump: skeleton for further writing the code
* goadifdxcc: add missing DXCC fields using godxcc
* goadifstat: obtain QSO statistics
* goadiftime: sort and filter QSOs by QSO\_DATE/TIME\_ON fields

## Things to do before compilation

```shell
go mod init github.com/jj1bdx/goadiftools
go mod tidy
```

## How to compile

Do `./buildall.sh`

## Required library

https://github.com/jj1bdx/adifparser
https://github.com/jj1bdx/godxcc

## License

BSD 2-clause License
