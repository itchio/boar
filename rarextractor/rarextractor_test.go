package rarextractor

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/itchio/boar/memfs"
	"github.com/itchio/dmcunrar-go/dmcunrar"
	"github.com/itchio/savior"
)

func TestEncryptedEntryPreservesSentinel(t *testing.T) {
	archivePath := generateEncryptedArchive(t)
	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("failed to read generated archive: %v", err)
	}

	ex, err := New(memfs.New(archiveBytes, "encrypted.rar"), nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = ex.Resume(nil, &savior.NopSink{})
	if err == nil {
		t.Fatal("Resume() error = nil, want encrypted error")
	}
	if !errors.Is(err, dmcunrar.ErrEncrypted) {
		t.Fatalf("Resume() error = %v, want dmcunrar.ErrEncrypted", err)
	}
}

func generateEncryptedArchive(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("rar"); err != nil {
		t.Skip("rar command not available")
	}

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "secret.txt")
	if err := os.WriteFile(sourcePath, []byte("secret\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	archivePath := filepath.Join(tmpDir, "encrypted.rar")
	cmd := exec.Command("rar", "a", "-idq", "-psecret", "encrypted.rar", "secret.txt")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rar failed: %v\n%s", err, output)
	}

	return archivePath
}
