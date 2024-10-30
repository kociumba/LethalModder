package ziputil

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// utils for adding to zip archives
type ZipWriter struct {
	Writer *zip.Writer
}

func (z *ZipWriter) WriteString(name, data string) error {
	w, err := z.Writer.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(data))
	return err
}

// extract a zip into target
func Extract(src io.ReaderAt, size int64, target string) error {
	reader, err := zip.NewReader(src, size)
	if err != nil {
		return err
	}

	// Ensure the target directory exists.
	if err := os.MkdirAll(target, os.ModePerm); err != nil {
		return err
	}

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue // skip directories; create them when copying files
		}

		relativePath := file.Name
		if filepath.Separator == '/' {
			relativePath = strings.ReplaceAll(relativePath, "\\", "/") // normalize on Unix
		}

		// Check for path traversal and skip if file escapes target
		if !isEnclosed(relativePath) {
			log.Printf("Skipping file %s: escapes archive root", relativePath)
			continue
		}

		outputPath := filepath.Join(target, relativePath)
		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			return err
		}

		if err := writeFile(file, outputPath); err != nil {
			return err
		}

		// Set Unix file permissions if on Unix
		if file.Mode() != 0 {
			_ = os.Chmod(outputPath, file.Mode())
		}
	}

	return nil
}

func writeFile(file *zip.File, outputPath string) error {
	inFile, err := file.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, inFile)
	return err
}

func isEnclosed(path string) bool {
	cleanPath := filepath.Clean(path)
	return !strings.HasPrefix(cleanPath, ".."+string(filepath.Separator))
}
