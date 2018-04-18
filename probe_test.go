package boar

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/itchio/wharf/state"
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
		{"foo_bar.rar", StrategySevenZip},
		{"foo_bar.dmg", StrategySevenZip},
		{"foo_bar.exe", StrategySevenZip},
		{"foo_bar", StrategySevenZip},
	}
)

func TestGetStrategy(t *testing.T) {
	consumer := &state.Consumer{}
	for _, cas := range strategyTests {
		ff := fakeFile{
			fileName: cas.fileName,
			canStat:  true,
		}
		strat := getStrategy(ff, consumer)
		assert.Equal(t, cas.result, strat)
	}
}

func TestGetStrategyNoStat(t *testing.T) {
	// Only one test case here
	ff := fakeFile{}
	strat := getStrategy(ff, &state.Consumer{})
	assert.Equal(t, StrategyNone, strat)
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
