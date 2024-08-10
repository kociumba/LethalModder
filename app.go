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
	"github.com/wailsapp/wails/v3/pkg/application"
)

const itemsPerPage = 10

var filteredListings []SimplePackageListing

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
	Version     string `json:"version"`
	Description string `json:"description"`
	URL         string `json:"url"`
	DownloadURL string `json:"download_url"`
	Icon        string `json:"icon"`
}

// App struct
//
// Deprecated: Left over from wailsv2, use the new DataService struct
type App struct {
	ctx context.Context
}

type DataService struct{}

// NewApp creates a new App application struct
//
// Deprecated: Left over from wailsv2, use the new DataService struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
//
// Deprecated: Left over from wailsv2, use the new DataService struct
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (d *DataService) Return10Simple(currentIndex int, direction Direction) []SimplePackageListing {
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
			Version:     listing.Versions[0].VersionNumber,
			Description: listing.Versions[0].Description,
			URL:         listing.PackageURL,
			DownloadURL: listing.Versions[0].DownloadURL,
			Icon:        listing.Versions[0].Icon,
		}
	}

	return simplifiedSubset
}

// Turns out the data is so big even on 10 entries that it crashed webview2 bridge
//
// # Do not use from frontend, results in a stack overflow
func (d *DataService) Return10Listings(currentIndex int, direction Direction) []api.PackageListing {
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

func (d *DataService) GetTotalItems() int {
	log.Info(len(packageListings))

	return len(packageListings)
}

func (d *DataService) Download(listing SimplePackageListing) (string, error) {
	fileURL := listing.DownloadURL
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
	defer os.Remove(tempFile.Name())

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
	outputDirName := fmt.Sprintf("%s_%s", listing.Name, listing.Version)
	extractedDir := filepath.Join(execDir, outputDirName)
	if err := os.Mkdir(extractedDir, os.ModePerm); err != nil {
		return "", err
	}

	// Extract the ZIP file
	err = extractZip(tempFile.Name(), extractedDir)
	if err != nil {
		return "", err
	}

	app.Events.Emit(&application.WailsEvent{
		Name: "downloadComplete",
		Data: extractedDir,
	})

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
//
// Deprecated: Use SimplePackageListing.DownloadURL
func (d *DataService) GetDownloadURL(listing api.PackageListing) string {
	return listing.Versions[0].DownloadURL
}

func (d *DataService) Return10WithSearch(currentIndex int, direction Direction, search string) []SimplePackageListing {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from panic:", "error", r)
			// Return an empty slice on error
			filteredListings = nil
		}
	}()

	filteredListings = d.FilterMods(search)

	var start, end int
	if direction == Next {
		start = currentIndex
		end = currentIndex + itemsPerPage
	} else {
		start = currentIndex - itemsPerPage
		end = currentIndex
	}

	if start < 0 {
		start = 0
	}
	if end > len(packageListings) {
		end = len(packageListings)
	}
	if start > len(packageListings) {
		start = len(packageListings)
	}

	log.Info(filteredListings[start:end])
	return filteredListings[start:end]
}

func (d *DataService) GetTotalItemsFiltered() int {
	return len(filteredListings)
}

// Unsung hero of the search function xd
func (d *DataService) FilterMods(search string) []SimplePackageListing {
	var filteredListings []SimplePackageListing
	for _, listing := range packageListings {
		if strings.Contains(strings.ToLower(listing.Name), strings.ToLower(search)) {
			simpleListing := SimplePackageListing{
				Name:        listing.Name,
				Version:     listing.Versions[0].VersionNumber,
				Description: listing.Versions[0].Description,
				URL:         listing.PackageURL,
				DownloadURL: listing.Versions[0].DownloadURL,
				Icon:        listing.Versions[0].Icon,
			}
			filteredListings = append(filteredListings, simpleListing)
		}
	}
	return filteredListings
}

// shitass function, still don't know why the event doesn't get picked up
// maby multi window stuff
func (d *DataService) GetIsLethalCompanyInstalled() bool {
	return IsLethalCompanyInstalled
}
