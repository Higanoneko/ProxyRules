package projectroot

import (
	"fmt"
	"os"
	"path/filepath"
)

func Find(start string) (string, error) {
	current := filepath.Clean(start)
	for {
		if isProjectRoot(current) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("project root not found from %s", start)
		}
		current = parent
	}
}

func isProjectRoot(path string) bool {
	requiredDirs := []string{
		filepath.Join(path, "Base"),
		filepath.Join(path, "Base", "Head"),
		filepath.Join(path, "Base", "Rules"),
	}

	for _, dir := range requiredDirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			return false
		}
	}

	return true
}
