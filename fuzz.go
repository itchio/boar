//+build gofuzz

package boar

import (
	"github.com/itchio/boar/memfs"
)

_dummyConsumer := &state.Consumer{}

func Fuzz(data []byte) int {
	file := memfs.New(data, "data")
	_, err := Probe(&ProbeParams{
		File: file,
		Consumer: dummyConsumer,
	})
	if err != nil {
		panic(err)
	}
	return 0
}
