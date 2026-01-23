package boar

import (
	"errors"

	"github.com/itchio/dash"
	"github.com/itchio/headway/state"
	"github.com/itchio/httpkit/eos"
	"github.com/itchio/savior"
)

var (
	ErrUnrecognizedArchiveType = errors.New("Unrecognized archive type")
)

type LoadFunc func(state any) error
type SaveFunc func(state any) error

type ExtractParams struct {
	File       eos.File
	StagePath  string
	OutputPath string

	Consumer *state.Consumer

	Load LoadFunc
	Save SaveFunc
}

type ProbeParams struct {
	File      eos.File
	Consumer  *state.Consumer
	Candidate *dash.Candidate
	OnEntries func(entries []*savior.Entry)
}

type Contents struct {
	Entries []*Entry
}

// Entry refers to a file entry in an archive
type Entry struct {
	Name             string
	UncompressedSize int64
}
