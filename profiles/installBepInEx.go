package profiles

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func InstallBepInEx(url, name, version, gamePath string) error {
	fileName := filepath.Base(url)

	// Get the executable directory
	exec, err := os.Executable()
	if err != nil {
		return err
	}

	execDir := filepath.Dir(exec)

	// Create a temporary file in the executable directory
	tempFile, err := os.CreateTemp(execDir, fileName)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name()) // Ensure the temp file is removed after the function ends

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return err
	}

	// Close the temp file before extraction
	if err := tempFile.Close(); err != nil {
		return err
	}

	// Extract the contents of the BepInExPack folder within the ZIP to the game path
	err = extractBepInExPack(tempFile.Name(), gamePath)
	if err != nil {
		return err
	}

	return nil
}

func extractBepInExPack(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// We are interested only in files inside the BepInExPack directory
		if !strings.HasPrefix(f.Name, "BepInExPack/") {
			continue
		}

		// Determine the relative path within the BepInExPack folder
		relativePath := strings.TrimPrefix(f.Name, "BepInExPack/")
		destPath := filepath.Join(dest, relativePath)

		// Create directories if needed
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// Create the file directly in the game path
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// Use a separate defer to close the file after copying
		defer destFile.Close()

		rc, err := f.Open()
		if err != nil {
			return err
		}

		// Use a separate defer to close the reader
		defer rc.Close()

		if _, err = io.Copy(destFile, rc); err != nil {
			return err
		}
	}

	return nil
}
