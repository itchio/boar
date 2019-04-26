# boar

[![build status](https://git.itch.ovh/itchio/boar/badges/master/build.svg)](https://git.itch.ovh/itchio/boar/commits/master)
[![codecov](https://codecov.io/gh/itchio/boar/branch/master/graph/badge.svg)](https://codecov.io/gh/itchio/boar)
[![Go Report Card](https://goreportcard.com/badge/github.com/itchio/boar)](https://goreportcard.com/report/github.com/itchio/boar)
[![GoDoc](https://godoc.org/github.com/itchio/boar?status.svg)](https://godoc.org/github.com/itchio/boar)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/itchio/boar/blob/master/LICENSE)

boar will sniff any archive, and maybe even let you extract it.

Under the hood, it uses:

  * <https://github.com/itchio/savior> - for resumable extraction
  * <https://github.com/itchio/dash> - for file sniffing
  * <https://github.com/itchio/sevenzip-go> - for the files we don't
    have a custom decompressor for

## License

boar's code itself is under MIT License, see `LICENSE` for details.

However:

  * `szextractor` uses 7-zip, which is licensed under LGPL and BSD 3-clause, see <https://www.7-zip.org/>
  * `rarextractor` uses dmc_unrar, which has LGPL and BSD clauses, see <https://github.com/DrMcCoy/dmc_unrar>
