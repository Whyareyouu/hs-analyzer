package architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedContractsHaveNoRuntimeDependencies(t *testing.T) {
	t.Parallel()

	files := goFilesUnder(t, filepath.Join("..", "..", "pkg", "contracts"))
	for _, path := range files {
		data := readFile(t, path)
		for _, forbidden := range []string{
			`"hsanalyzer/internal/`,
			`"hsanalyzer/apps/`,
			`"hsanalyzer/cmd/`,
		} {
			if strings.Contains(data, forbidden) {
				t.Fatalf("shared contract %s imports forbidden runtime dependency %s", path, forbidden)
			}
		}
	}
}

func TestBackendDoesNotImportCLIRuntimeInternals(t *testing.T) {
	t.Parallel()

	files := goFilesUnder(t, filepath.Join("..", "..", "apps", "hs-analyzer-be"))
	for _, path := range files {
		data := readFile(t, path)
		for _, forbidden := range []string{
			`"hsanalyzer/cmd/hsanalyzer`,
			`"hsanalyzer/apps/hs-analyzer-cli`,
			`"hsanalyzer/internal/parser`,
			`"hsanalyzer/internal/logreader`,
			`"hsanalyzer/internal/monitor`,
		} {
			if strings.Contains(data, forbidden) {
				t.Fatalf("backend file %s imports CLI runtime dependency %s", path, forbidden)
			}
		}
	}
}

func goFilesUnder(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return files
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
