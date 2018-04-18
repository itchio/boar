//+build gofuzz

package boar

import (
	"github.com/itchio/boar/memfs"
)

_dummyConsumer := &state.Consumer{}

func Fuzz(data []byte) int {
	file := memfs.New(data, "data")
	params := &ProbeParams{
		File: file,
		Consumer: dummyConsumer,
	}

	if _, err := Probe(params); err != nil {
		return 0
	}
	return 1
}
