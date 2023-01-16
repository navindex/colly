package filesys

import (
	"os"
	"path/filepath"
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
