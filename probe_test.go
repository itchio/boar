package boar

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/itchio/headway/state"
	"github.com/itchio/httpkit/eos"
	"github.com/stretchr/testify/assert"
)

type StrategyTest struct {
	fileName string
	result   Strategy
}

var (
	strategyTests = []StrategyTest{
		{"foo_bar.zip", StrategyZip},
		{"foo_bar.tar", StrategyTar},
		{"foo_bar.tar.gz", StrategyTarGz},
		{"foo_bar.tar.bz2", StrategyTarBz2},
		{"foo_bar.7z", StrategySevenZip},
		{"foo_bar.exe", StrategySevenZipUnsure},
		{"foo_bar", StrategySevenZipUnsure},
	}
)

func TestGetStrategy(t *testing.T) {
	consumer := &state.Consumer{}
	for _, cas := range strategyTests {
		ff := fakeFile{
			fileName: cas.fileName,
			canStat:  true,
		}
		strat := getStrategy(getExt(ff, consumer))
		assert.Equal(t, cas.result, strat)
	}
}

func TestGetStrategyNoStat(t *testing.T) {
	// Only one test case here
	ff := fakeFile{}
	strat := getStrategy(getExt(ff, &state.Consumer{}))
	assert.Equal(t, StrategySevenZipUnsure, strat)
}

type fakeFile struct {
	fileName string
	canStat  bool
}

func (ff fakeFile) Read([]byte) (int, error) {
	return 0, errors.New("Fake files can't read")
}
func (ff fakeFile) Close() error {
	return errors.New("Fake files can't close")
}
func (ff fakeFile) ReadAt([]byte, int64) (int, error) {
	return 0, errors.New("Fake files can't read")
}
func (ff fakeFile) Seek(int64, int) (int64, error) {
	return 0, errors.New("Fake files can't seek")
}
func (ff fakeFile) Stat() (os.FileInfo, error) {
	if ff.canStat {
		return fakeFileInfo{name: ff.fileName}, nil
	}
	return fakeFileInfo{}, errors.New("This fake file can't Stat()")
}

type fakeFileInfo struct {
	name string
}

func (ffi fakeFileInfo) Name() string {
	return ffi.name
}
func (ffi fakeFileInfo) Size() int64 {
	return 0
}
func (ffi fakeFileInfo) IsDir() bool {
	return false
}
func (ffi fakeFileInfo) ModTime() time.Time {
	return time.Time{}
}
func (ffi fakeFileInfo) Mode() os.FileMode {
	return 0
}
func (ffi fakeFileInfo) Sys() interface{} {
	return nil
}

func Test_RealFiles(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		panic(err)
	}

	consumer := &state.Consumer{}
	for _, f := range files {
		t.Run(f.Name(), func(t *testing.T) {
			file, err := eos.Open(filepath.Join("testdata", f.Name()))
			if err != nil {
				panic(err)
			}

			defer file.Close()
			ai, err := Probe(ProbeParams{
				File:     file,
				Consumer: consumer,
			})
			if err != nil {
				panic(err)
			}

			tokens := strings.Split(f.Name(), "-")
			lastToken := tokens[len(tokens)-1]

			expectedType := lastToken
			switch lastToken {
			case "rar4":
				expectedType = "rar"
			case "gz":
				expectedType = "tar.gz"
			case "bz2":
				expectedType = "tar.bz2"
			case "xz":
				expectedType = "tar.xz"
			}

			assert.EqualValues(t, expectedType, ai.Format)
		})
	}
}
