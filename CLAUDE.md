# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build

- Build everything: `./buildall.sh` (finds each `goadif*` and `noasciitostar` directory and runs `go build` inside).
- Build one tool: `cd <tool-dir> && go build`.
- After cloning fresh / changing dependencies: `go mod tidy`. Module path is `github.com/jj1bdx/goadiftools`; the three external libraries (`adifparser`, `gocldb`, `godxcc`) are all under `github.com/jj1bdx`.
- There are no Go tests in this repo — `go test ./...` does nothing meaningful.

## Architecture

This repository is a flat collection of independent command-line tools, not a library. Each subdirectory is its own `package main` with a single `.go` file of the same name as the directory; the produced binary sits next to the source. There is no shared internal package — code that recurs (flag parsing, file open/create, refuse-to-overwrite check, ADIF read loop) is duplicated by design across tools. When making cross-cutting changes, expect to touch every tool's `main.go`.

### Unix-filter convention

Every tool follows the same I/O contract, which is what lets them be piped together (see `examples/wsjtxlog-adifgen.sh`):

- `-f infile` — input file; stdin if omitted.
- `-o outfile` — output file; stdout if omitted. **Refuses to overwrite an existing file** (checked via `os.Stat` + `os.IsNotExist`); this is intentional safety, not a bug.
- Errors are written to stderr; the tool continues or returns. EOF breaks the read loop cleanly.

When adding a new tool, copy this scaffolding from an existing one (e.g. `goadifcsv/goadifcsv.go`) rather than inventing a new convention.

### ADIF processing pattern

Tools that read/write ADIF use `github.com/jj1bdx/adifparser`:

```
reader := adifparser.NewADIFReader(fp)
writer := adifparser.NewADIFWriter(writefp)
for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
    // record.GetValue("call") returns adifparser.ErrNoSuchField for missing fields — handle this sentinel explicitly
    // record.SetValue("country", ...) to add/overwrite
    writer.WriteRecord(record)
}
writer.Flush()
```

ADIF field names are lowercase by convention in this codebase (`record.GetValue(strings.ToLower(name))`). For deduping, swap `NewADIFReader` for `NewDedupeADIFReader` — same interface.

`noasciitostar` is the exception: it operates on raw UTF-8 bytes via `bufio`, not ADIF records. Treat it as a plain text filter that happens to live in the same repo.

### DXCC enrichment

`goadifdxcc` uses `godxcc` (callsign-only lookup); `goadifdxcccl` uses `gocldb` (Club Log database, time-aware — needs `QSO_DATE`/`TIME_ON` parsed into a `time.Time`). When in doubt, prefer `goadifdxcccl` because Club Log handles historical prefix changes. `gocldb.LoadCtyXml()` must be called once at startup; `gocldb.SetDebugOutput(io.Discard)` silences its logging. Note: `gocldb` has no ITU Zone data — only DXCC/CQZ/continent.

### Field-fill semantics in DXCC tools

Both DXCC tools only fill a field when `GetValue` returns `ErrNoSuchField` — they do **not** overwrite existing values. CQZ is additionally guarded against zero (`result.Cqz > 0`) because zero is not a valid CQ Zone. Preserve this "fill if empty, never overwrite" behavior when extending these tools.
