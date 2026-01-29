# boar

[![build](https://github.com/itchio/boar/actions/workflows/test.yml/badge.svg)](https://github.com/itchio/boar/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/itchio/boar)](https://goreportcard.com/report/github.com/itchio/boar)
[![GoDoc](https://godoc.org/github.com/itchio/boar?status.svg)](https://godoc.org/github.com/itchio/boar)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/itchio/boar/blob/master/LICENSE)

boar will sniff any archive, and maybe even let you extract it.

## Format Support

| Format | Extractor | Native Dependency |
|--------|-----------|-------------------|
| `.zip` | `zipextractor` (via [savior](https://github.com/itchio/savior)) | None (pure Go) |
| `.tar` | `tarextractor` (via [savior](https://github.com/itchio/savior)) | None (pure Go) |
| `.tar.gz` | `tarextractor` + `gzipsource` | None (pure Go) |
| `.tar.bz2` | `tarextractor` + `bzip2source` | None (pure Go) |
| `.tar.xz` | `tarextractor` + `xzsource` | Requires 7-zip libs |
| `.rar` | `rarextractor` (via [dmcunrar-go](https://github.com/itchio/dmcunrar-go)) | `dmc_unrar` (C library) |
| `.7z` | `szextractor` (via [sevenzip-go](https://github.com/itchio/sevenzip-go)) | `7z.so` + `libc7zip.so` |
| `.exe` (self-extracting) | `szextractor` | `7z.so` + `libc7zip.so` |

Formats with native dependencies will fail to extract if the required libraries are not available. The pure Go extractors work without any additional dependencies.

## Dependencies

  * <https://github.com/itchio/savior> - for resumable extraction (zip, tar)
  * <https://github.com/itchio/dash> - for file sniffing
  * <https://github.com/itchio/sevenzip-go> - for 7z and self-extracting archives
  * <https://github.com/itchio/dmcunrar-go> - for RAR archives

## License

boar's code itself is under MIT License, see `LICENSE` for details.

However:

  * `szextractor` uses 7-zip, which is licensed under LGPL and BSD 3-clause, see <https://www.7-zip.org/>
  * `rarextractor` uses dmc_unrar, which has LGPL and BSD clauses, see <https://github.com/DrMcCoy/dmc_unrar>
