package szextractor_test

import (
	"log"
	"testing"

	"github.com/itchio/boar/memfs"
	"github.com/itchio/boar/szextractor"
	"github.com/itchio/headway/state"
	"github.com/itchio/savior"
	"github.com/itchio/savior/checker"
	"github.com/stretchr/testify/assert"
)

func must(t *testing.T, err error) {
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
}

func TestSzExtractor(t *testing.T) {
	sink := checker.MakeTestSinkAdvanced(40)
	zipBytes := checker.MakeZip(t, sink)

	file := memfs.New(zipBytes, "szextractor_test.zip")

	initialConsumer := &state.Consumer{
		OnMessage: func(lvl string, message string) {
			log.Printf("[%s] %s", lvl, message)
		},
	}

	makeExtractor := func() savior.Extractor {
		ex, err := szextractor.New(file, initialConsumer)
		must(t, err)
		return ex
	}

	log.Printf("Testing szextractor on .zip, no resumes")
	checker.RunExtractorText(t, makeExtractor, sink, func() bool {
		return false
	})

	log.Printf("Testing szextractor on .zip, all resumes")
	checker.RunExtractorText(t, makeExtractor, sink, func() bool {
		return true
	})

	log.Printf("Testing szextractor on .zip, every other")
	i := 0
	checker.RunExtractorText(t, makeExtractor, sink, func() bool {
		i++
		return i%2 == 0
	})
}
