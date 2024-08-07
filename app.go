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

	"github.com/charmbracelet/log"
	"github.com/kociumba/LethalModder/api"
)

const itemsPerPage = 10

type Direction int

const (
	Next Direction = iota
	Previous
)

// Used to make sure webview2 bridge doesn't get overloaded.
//
// Has to be json tagged to translate into JS
type SimplePackageListing struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	DownloadURL string `json:"download_url"`
}

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

func (a *App) Return10Simple(currentIndex int, direction Direction) []SimplePackageListing {
	var start, end int
	if direction == Next {
		start = currentIndex
		end = currentIndex + itemsPerPage
	} else {
		start = currentIndex - itemsPerPage
		end = currentIndex
	}

	// Ensure start and end are within valid bounds
	if start < 0 {
		start = 0
	}
	if end > len(packageListings) {
		end = len(packageListings)
	}
	if start > len(packageListings) {
		start = len(packageListings)
	}

	subset := packageListings[start:end]
	simplifiedSubset := make([]SimplePackageListing, len(subset))
	for i, listing := range subset {
		simplifiedSubset[i] = SimplePackageListing{
			Name:        listing.Name,
			Description: listing.Versions[0].Description,
			URL:         listing.PackageURL,
			DownloadURL: listing.Versions[0].DownloadURL,
		}
	}

	return simplifiedSubset
}

// Turns out the data is so big even on 10 entries that it crashed webview2 bridge
//
// # Do not use from frontend, results in a stack overflow
func (a *App) Return10Listings(currentIndex int, direction Direction) []api.PackageListing {
	var start, end int
	if direction == Next {
		start = currentIndex
		end = currentIndex + itemsPerPage
	} else {
		start = currentIndex - itemsPerPage
		end = currentIndex
	}

	// Ensure start and end are within valid bounds
	if start < 0 {
		start = 0
	}
	if end > len(packageListings) {
		end = len(packageListings)
	}
	if start > len(packageListings) {
		start = len(packageListings)
	}

	log.Info(packageListings[start:end])
	return packageListings[start:end]
}

func (a *App) GetTotalItems() int {
	log.Info(len(packageListings))

	return len(packageListings)
}

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
// Deprecated as I already return this in the simplified package listing
func (a *App) GetDownloadURL(listing api.PackageListing) string {
	return listing.Versions[0].DownloadURL
}
