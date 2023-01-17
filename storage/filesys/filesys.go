package filesys

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kennygrant/sanitize"
)

// ------------------------------------------------------------------------

// FileCount returns the number of files in the directory and its subdirectories.
func FileCount(path string) (count uint, err error) {
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == "" {
			count++
		}
		return err
	})
	if err != nil {
		count = 0
	}

	return count, err
}

// ------------------------------------------------------------------------

// SanitizeFileName replaces dangerous characters in a string
// so the return value can be used as a safe file name.
func SanitizeFileName(fileName string) string {
	ext := sanitize.BaseName(filepath.Ext(fileName))
	name := sanitize.BaseName(fileName[:len(fileName)-len(ext)])

	if ext == "" {
		ext = ".unknown"
	}

	return strings.Replace(name+ext, "-", "_", -1)
}
