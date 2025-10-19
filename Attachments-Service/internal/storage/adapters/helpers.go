package adapters

import (
	"io"
	"os"
)

// OpenFile opens a file for reading
func OpenFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
