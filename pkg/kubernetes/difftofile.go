package kubernetes

import (
	"github.com/pkg/errors"
	"os"
)

// always create file for consistency
func writeDiffToFile(path string, diff *string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if diff != nil {
		_, err = f.WriteString(*diff)
		if err != nil {
			return err
		}
	}
	return f.Close()
}

func WriteDiffToFile(path string, diff *string) error {
	err := writeDiffToFile(path, diff)
	if err != nil {
		return errors.Wrap(err, "writting diff to file")
	}
	return nil
}

// helper to produce and empty file
// when we get empty diff we should produce an empty file
func TruncateDiffToFile(path string) error {
	return WriteDiffToFile(path, nil)
}
