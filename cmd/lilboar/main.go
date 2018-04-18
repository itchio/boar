package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/itchio/boar"
	"github.com/itchio/wharf/eos"
	"github.com/itchio/wharf/state"
)

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("Usage: lilboar FILE [...FILE]")
	}

	consumer := &state.Consumer{
		OnMessage: func(lvl string, msg string) {
			log.Printf("%s", msg)
		},
	}

	doFile := func(filePath string) {
		file, err := eos.Open(filePath)
		if err != nil {
			consumer.Errorf("%s: %v", filePath, err)
			return
		}
		defer file.Close()

		info, err := boar.Probe(&boar.ProbeParams{
			File:     file,
			Consumer: consumer,
		})
		if err != nil {
			consumer.Errorf("%s: %v", filePath, err)
			return
		}

		consumer.Infof("%s: %s", filepath.Base(filePath), info)
	}

	for _, arg := range args {
		doFile(arg)
	}
}
