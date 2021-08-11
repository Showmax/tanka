package kubernetes

import (
	"github.com/pkg/errors"
	"os"
)

// Write string to file
// Create empty file if passed string is nil
func writeStringToFile(path string, str *string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if str != nil {
		_, err = f.WriteString(*str)
		if err != nil {
			return err
		}
	}
	return f.Close()
}

func WriteDiffToFile(path string, diff *string) error {
	err := writeStringToFile(path, diff)
	if err != nil {
		return errors.Wrap(err, "writting diff to file")
	}
	return nil
}

// Helper to produce an empty file
// When we get empty diff we should produce an empty file
func TruncateDiffFile(path string) error {
	return WriteDiffToFile(path, nil)
}
