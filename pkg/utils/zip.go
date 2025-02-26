package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"
)

func UnzipZipFile(in []byte) (map[string][]byte, error) {
	reader := bytes.NewReader(in)
	zipReader, err := zip.NewReader(reader, int64(len(in)))
	if err != nil {
		return nil, err
	}

	filesContent := make(map[string][]byte)
	for _, file := range zipReader.File {
		err := func() error {
			f, err := file.Open()
			if err != nil {
				return err
			}
			defer func(f io.ReadCloser) {
				err := f.Close()
				if err != nil {
					slog.Error(fmt.Sprintf("close file: %v", err))
				}
			}(f)

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, f); err != nil {
				return err
			}

			filesContent[file.Name] = buf.Bytes()

			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return filesContent, nil // Example usage
}
