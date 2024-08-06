package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kociumba/LethalModder/api"
)

const itemsPerPage = 10

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Return10Listings(currentIndex, direction int) []api.PackageListing {
	totalItems := len(packageListings)

	// Handle out of bounds
	if currentIndex < 0 {
		currentIndex = 0
	} else if currentIndex >= totalItems {
		currentIndex = totalItems - itemsPerPage
		if currentIndex < 0 {
			currentIndex = 0
		}
	}

	var listings []api.PackageListing
	if direction > 0 {
		endIndex := currentIndex + itemsPerPage
		if endIndex > totalItems {
			endIndex = totalItems
		}
		listings = packageListings[currentIndex:endIndex]
	} else {
		startIndex := currentIndex - itemsPerPage
		if startIndex < 0 {
			startIndex = 0
		}
		listings = packageListings[startIndex:currentIndex]
	}

	return listings
}

func (a *App) GetTotalItems() int {
	return len(packageListings)
}

// func (a *App) OpenWebsite(url string) {
// 	// Open a website in the system default browser
// 	var cmd string
// 	var args []string

// 	switch runtime.GOOS {
// 	case "windows":
// 		cmd = "cmd"
// 		args = []string{"/c", "start"}
// 	case "darwin":
// 		cmd = "open"
// 	default: // "linux", "freebsd", "openbsd", "netbsd"
// 		cmd = "xdg-open"
// 	}

// 	args = append(args, url)
// 	exec.Command(cmd, args...).Start()
// }

func (a *App) Download(url string) (string, error) {
	fileURL := url
	fileName := filepath.Base(fileURL)

	exec, err := os.Executable()
	if err != nil {
		return "", err
	}

	execDir := filepath.Dir(exec)

	// Create a temporary file
	tempFile, err := os.CreateTemp(execDir, fileName)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name()) // Remove temp file after extraction

	// Download the file
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: %s", resp.Status)
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	// Close the temp file before extraction
	if err := tempFile.Close(); err != nil {
		return "", err
	}

	// Create a new folder for the extracted contents
	extractedDir := filepath.Join(execDir, strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	if err := os.Mkdir(extractedDir, os.ModePerm); err != nil {
		return "", err
	}

	// Extract the ZIP file
	err = extractZip(tempFile.Name(), extractedDir)
	if err != nil {
		return "", err
	}

	return extractedDir, nil
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		destPath := filepath.Join(dest, f.Name)

		// Create directories if needed
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// Create the file
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

// [0] is the newest
func (a *App) GetDownloadURL(listing api.PackageListing) string {
	return listing.Versions[0].DownloadURL
}
