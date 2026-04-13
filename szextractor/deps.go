package szextractor

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/itchio/arkive/zip"
	"github.com/itchio/boar/szextractor/formulas"
	"github.com/itchio/boar/szextractor/types"
	"github.com/itchio/headway/state"
	"github.com/itchio/headway/united"
	"github.com/itchio/httpkit/eos"
	"github.com/itchio/httpkit/eos/option"
	"github.com/nightlyone/lockfile"
	"github.com/pkg/errors"
)

func getDepSpec() *types.DepSpec {
	osarch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	if ds, ok := formulas.ByOsArch[osarch]; ok {
		return &ds
	}

	return nil
}

type tempLockfileErr interface {
	Temporary() bool
}

func depChannel() string {
	return fmt.Sprintf("%s-%s-%s", runtime.GOOS, runtime.GOARCH, depChannelOverride)
}

func channelSourceURL(channel string) string {
	return fmt.Sprintf("https://broth.itch.zone/libc7zip/%s/LATEST/archive.zip", channel)
}

func acquireDepLock(consumer *state.Consumer, execDir string) (lockfile.Lockfile, error) {
	lockFilePath := filepath.Join(execDir, ".boar-deps.lock")
	lf, err := lockfile.New(lockFilePath)
	if err != nil {
		return "", errors.WithStack(err)
	}

	err = lf.TryLock()
	tries := 10
	for err != nil {
		if err == lockfile.ErrBusy {
			time.Sleep(2 * time.Second)
			err = lf.TryLock()
			continue
		}

		// lockfile's recommended way to look for a temporary error
		if _, ok := err.(tempLockfileErr); ok {
			consumer.Debugf("Will retry acquiring lock in a few: %s", err.Error())
			tries--
			if tries <= 0 {
				msg := fmt.Sprintf("Too many errors acquiring lock, giving up. Last error was: %s", err.Error())
				return "", errors.New(msg)
			}

			time.Sleep(2 * time.Second)
		} else {
			return "", errors.WithStack(err)
		}
	}

	return lf, nil
}

// InstallDeps unconditionally downloads and installs the native 7-zip
// libraries, regardless of whether they already exist on disk. Uses the
// channel override if set via SetDepChannel, otherwise uses the default
// formula source.
func InstallDeps(consumer *state.Consumer) error {
	depSpec := getDepSpec()
	if depSpec == nil {
		return errors.Errorf("No dependencies available for %s-%s", runtime.GOOS, runtime.GOARCH)
	}

	execPath, err := os.Executable()
	if err != nil {
		return errors.WithStack(err)
	}
	execDir := filepath.Dir(execPath)

	lf, err := acquireDepLock(consumer, execDir)
	if err != nil {
		return err
	}
	defer lf.Unlock()

	var source string
	if depChannelOverride != "" {
		channel := depChannel()
		source = channelSourceURL(channel)
		consumer.Opf("Installing dependencies from %s", source)
	} else {
		if len(depSpec.Sources) == 0 {
			return errors.New("No sources available for dependencies")
		}
		source = depSpec.Sources[0]
		consumer.Opf("Installing dependencies from %s", source)
	}

	err = fetchDeps(consumer, source, depSpec.Entries, execDir)
	if err != nil {
		return err
	}

	ensuredDeps = true
	return nil
}

func EnsureDeps(consumer *state.Consumer) error {
	if dontEnsureDeps {
		consumer.Debugf("Asked not to ensure dependencies, skipping...")
		return nil
	}

	if ensuredDeps {
		return nil
	}

	consumer.Debugf("Ensuring dependencies...")
	depSpec := getDepSpec()
	if depSpec == nil {
		consumer.Debugf("No dependencies for %s-%s", runtime.GOOS, runtime.GOARCH)
		return nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return errors.WithStack(err)
	}
	execDir := filepath.Dir(execPath)

	lf, err := acquireDepLock(consumer, execDir)
	if err != nil {
		return err
	}
	defer lf.Unlock()

	if depChannelOverride != "" {
		err = ensureChannelDeps(consumer, depSpec, execDir)
	} else {
		err = ensureFormulaDeps(consumer, depSpec, execDir)
	}
	if err != nil {
		return err
	}

	ensuredDeps = true
	return nil
}

func ensureChannelDeps(consumer *state.Consumer, depSpec *types.DepSpec, execDir string) error {
	channel := depChannel()

	// Check which files are missing (no hash verification for channel deps)
	var toFetch []types.DepEntry
	for _, entry := range depSpec.Entries {
		entryPath := filepath.Join(execDir, entry.Name)
		if _, err := os.Stat(entryPath); err != nil {
			consumer.Debugf("[%s] missing, will fetch from channel %s", entry.Name, channel)
			toFetch = append(toFetch, entry)
		}
	}

	if len(toFetch) == 0 {
		consumer.Debugf("All dependencies present for channel %s, skipping fetch", channel)
		return nil
	}

	source := channelSourceURL(channel)
	consumer.Opf("Fetching dependencies from channel %s...", channel)

	return fetchDeps(consumer, source, toFetch, execDir)
}

func ensureFormulaDeps(consumer *state.Consumer, depSpec *types.DepSpec, execDir string) error {
	var toFetch []types.DepEntry

	for _, entry := range depSpec.Entries {
		func() {
			entryPath := filepath.Join(execDir, entry.Name)

			f, err := os.Open(entryPath)
			if err != nil {
				consumer.Debugf("[%s] could not open, will fetch", entry.Name)
				if !os.IsNotExist(err) {
					consumer.Debugf("  %s", err.Error())
				}
				toFetch = append(toFetch, entry)
				return
			}

			hashes := make(map[types.HashAlgo]hash.Hash)
			for _, dh := range entry.Hashes {
				switch dh.Algo {
				case types.HashAlgoSHA1:
					hashes[dh.Algo] = sha1.New()
				case types.HashAlgoSHA256:
					hashes[dh.Algo] = sha256.New()
				}
			}

			if len(hashes) == 0 {
				consumer.Debugf("No hashes to check, calling it a day.")
				return
			}

			// oh go, I'm so disappoint in you right now
			var writers []io.Writer
			for _, h := range hashes {
				writers = append(writers, h)
			}
			mw := io.MultiWriter(writers...)

			_, err = io.Copy(mw, f)
			if err != nil {
				consumer.Debugf("Error during hashing of %s, will fetch: %s", entry.Name, err.Error())
				toFetch = append(toFetch, entry)
				return
			}

			for _, dh := range entry.Hashes {
				h := hashes[dh.Algo]
				if h != nil {
					expected := dh.Value
					// yes, yes, bytes.Equal is a thing. but also
					// []byte{} literals are not the friendliest. don't @ me.
					actual := fmt.Sprintf("%x", h.Sum(nil))
					if actual != expected {
						consumer.Debugf("[%s] %s hash mismatch, will fetch", entry.Name, dh.Algo)
						consumer.Debugf("  wanted %s", expected)
						consumer.Debugf("     got %s", actual)
						toFetch = append(toFetch, entry)
						return
					}
				}
			}
		}()
	}

	if len(toFetch) > 0 {
		consumer.Logf("")
		consumer.Opf("Healing %d dependencies...", len(toFetch))

		firstSource := true
		for _, source := range depSpec.Sources {
			if !firstSource {
				consumer.Logf("Trying next source...")
			}

			firstSource = false
			err := fetchDeps(consumer, source, toFetch, execDir)
			if err != nil {
				consumer.Logf("Error while installing dependencies: %s", err.Error())
				continue
			}
			break
		}
		consumer.Logf("")
	}

	return nil
}

func fetchDeps(consumer *state.Consumer, source string, toFetch []types.DepEntry, execDir string) error {
	beforeHeal := time.Now()

	f, err := eos.Open(source, option.WithConsumer(consumer))
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	stats, err := f.Stat()
	if err != nil {
		return errors.WithStack(err)
	}

	zr, err := zip.NewReader(f, stats.Size())
	if err != nil {
		return errors.WithStack(err)
	}

	foundFiles := 0
	var installedSize int64
	for _, zf := range zr.File {
		for _, entry := range toFetch {
			if entry.Name == zf.Name {
				foundFiles++
				consumer.Opf("%s (%s)...", entry.Name, united.FormatBytes(int64(zf.UncompressedSize64)))
				entryPath := filepath.Join(execDir, entry.Name)

				err = func() error {
					zer, err := zf.Open()
					if err != nil {
						return errors.WithStack(err)
					}
					defer zer.Close()

					of, err := os.Create(entryPath)
					if err != nil {
						return errors.WithStack(err)
					}
					defer of.Close()

					writtenBytes, err := io.Copy(of, zer)
					if err != nil {
						return errors.WithStack(err)
					}

					installedSize += writtenBytes
					return nil
				}()

				if err != nil {
					return errors.WithStack(err)
				}
			}
		}
	}

	if foundFiles < len(toFetch) {
		return errors.Errorf("Found only %d files of the required %d", foundFiles, len(toFetch))
	}
	consumer.Statf("Installed %s's worth of dependencies in %s", united.FormatBytes(installedSize), time.Since(beforeHeal))

	return nil
}
