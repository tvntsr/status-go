// +build wasm

package fileutil


// CopyDirs copies all directories, subdirectories and files recursively including the empty folders.
// Source and destination must be full paths.
func CopyDirs(src, dest string) error {
	return nil
}

func copyFile(src, dest string) error {
	return nil
}

// readDirs reads the source directory recursively and
// returns relative paths to all files and empty directories.
func readDirs(src string) ([]string, error) {
	return []string{}, nil
}

// ReadDir returns the filenames in the given directory in sorted order.
func ReadDir(dirpath string) ([]string, error) {
	return []string{}, nil
}

// Rename safely renames a file.
func Rename(from, to string) error {
	return nil
}

// Replace moves a file or directory to a new location and deletes any previous data.
// It is not atomic.
func Replace(from, to string) error {
	return nil
}
