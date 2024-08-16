package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/kociumba/LethalModder/api"
	"github.com/kociumba/LethalModder/profiles"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const itemsPerPage = 10

var filteredListings []SimplePackageListing

type Direction int

const (
	Next Direction = iota
	Previous
)

// var packageMap = make(map[string]api.PackageListing) // still not sure if this is a good idea

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

type ModManifest struct {
	Name          string   `json:"name"`
	Author        string   `json:"author"`
	VersionNumber string   `json:"version_number"`
	WebsiteURL    string   `json:"website_url"`
	Description   string   `json:"description"`
	Dependencies  []string `json:"dependencies"`
}

// The new main data service for the frontend
// All the old code from wailsv2 was deleted.
type DataService struct {
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

func InitializeGlobalMaps(packageListings []api.PackageListing) {
	globalMu.Lock()
	defer globalMu.Unlock()

	packageMap = make(map[string]api.PackageListing, len(packageListings))
	for _, listing := range packageListings {
		packageMap[listing.Name] = listing
	}
	downloadedMods = make(map[string]bool)
}

// overcomplicated, r2modman just extracts everything from the bepinex
func (d *DataService) Download(listing SimplePackageListing) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from panic:", "error", r)
			// Return an empty slice on error
			filteredListings = nil
		}
	}()

	globalMu.Lock()
	if downloadedMods[listing.Name] {
		globalMu.Unlock()
		return "", nil // Mod already downloaded
	}
	downloadedMods[listing.Name] = true
	globalMu.Unlock()

	fileURL := listing.DownloadURL
	fileName := filepath.Base(fileURL)

	// Use the selected profile
	if SelectProfile.Path == "" {
		log.Error("no profile selected")
		return "", fmt.Errorf("no profile selected")
	}

	bepInExDir := filepath.Join(SelectProfile.Path, "BepInEx")

	tempFile, err := os.CreateTemp("", fileName)
	if err != nil {
		log.Error(err)
		return "", err
	}
	defer os.Remove(tempFile.Name())

	if err := d.downloadFile(fileURL, tempFile); err != nil {
		log.Error(err)
		return "", err
	}

	extractedDir, manifest, err := d.extractAndParseManifest(tempFile.Name(), bepInExDir, listing.Name, listing.Version)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if err := d.handleDependencies(manifest, bepInExDir); err != nil {
		log.Error(err)
		return "", err
	}

	app.Events.Emit(&application.WailsEvent{
		Name: "downloadComplete",
		Data: extractedDir,
	})

	return extractedDir, nil
}

func (d *DataService) downloadFile(url string, file *os.File) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		app.Events.Emit(&application.WailsEvent{
			Name: "downloadFailed",
			Data: resp.Status,
		})

		log.Error("Failed to download file: " + resp.Status)
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

// Extraction needs to account for the insane inconsistency in the way mods are packaged
//
//   - some contain BepInEx/plugins/mod
//   - some contain BepInEx/many bepinex dirs/mods or configs
//   - some contain just the mod in the root dir
//   - some contain only a manifest with deps (modpacks)
//
// all the mods have a manifest and we need to recursively download the mods from deps listed there
//
// In each case the default behavior should be to place the whole downloaded mod into BepInEx/plugins/modDir after unzipping
//
// When we have a BepInEx/... structure in the download we need to mimic the structure in our profile dir and place unzipped files in corresponding dirs.
// This should be able to handle many dirs inside BepInEx/ as sometimes there are many dirs with different contents.
// When mimicking the structure dirs need to be created in BepInEx/ as not always the expected dirs will already exist.
//
// Before doing any of that we need to read the manifest.json into
//
//	ModManifest{}
//
// and recursively call
//
//	Download()
//
// to get all the mods required by all of the mods.
func (d *DataService) extractAndParseManifest(src, dest, modName, modVersion string) (string, *ModManifest, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from panic:", "error", r)
			// Return an empty slice on error
			filteredListings = nil
		}
	}()

	r, err := zip.OpenReader(src)
	if err != nil {
		log.Error(err)
		return "", nil, err
	}
	defer r.Close()

	extractedDir := filepath.Join(dest, "plugins", fmt.Sprintf("%s_%s", modName, modVersion))
	var manifestFile *zip.File

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "BepInEx/") {
			// Handle BepInEx structure
			destPath := filepath.Join(dest, f.Name)
			if err := d.extractFile(f, destPath); err != nil {
				log.Error(err)
				return "", nil, err
			}
			extractedDir = filepath.Dir(destPath)
		} else if filepath.Base(f.Name) == "manifest.json" {
			manifestFile = f
		} else {
			// Handle root directory mods
			destPath := filepath.Join(extractedDir, f.Name)
			if err := d.extractFile(f, destPath); err != nil {
				log.Error(err)
				return "", nil, err
			}
		}
	}

	if manifestFile == nil {
		log.Error("manifest.json not found")
		return "", nil, fmt.Errorf("manifest.json not found")
	}

	manifest, err := d.parseManifest(manifestFile)
	if err != nil {
		log.Error(err)
		return "", nil, err
	}

	return extractedDir, manifest, nil
}

func (d *DataService) extractFile(f *zip.File, destPath string) error {
	if f.FileInfo().IsDir() {
		return os.MkdirAll(destPath, os.ModePerm)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		log.Error(err)
		return err
	}

	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		log.Error(err)
		return err
	}
	defer destFile.Close()

	srcFile, err := f.Open()
	if err != nil {
		log.Error(err)
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (d *DataService) parseManifest(file *zip.File) (*ModManifest, error) {
	rc, err := file.Open()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		log.Error(err)
		return nil, err
	}

	// Strip BOM if present
	manifestBytes := buf.Bytes()
	if len(manifestBytes) > 0 && manifestBytes[0] == 0xEF && manifestBytes[1] == 0xBB && manifestBytes[2] == 0xBF {
		manifestBytes = manifestBytes[3:] // Remove the BOM
	}

	var manifest ModManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		log.Error("fucked up unmarshaling manifest: ", "error", err, "manifest", buf.String())
		return nil, err
	}

	log.Info("manifest: ", "manifest", manifest)

	return &manifest, nil
}

// the second arg is BepInExDir, not used for now
func (d *DataService) handleDependencies(manifest *ModManifest, _ string) error {
	filteredDeps := d.filterDependencies(manifest.Dependencies)
	depListings, err := d.getModListingsForDependencies(filteredDeps)
	if err != nil {
		log.Errorf("failed to get listings for dependencies: %v", err)
		return fmt.Errorf("failed to get listings for dependencies: %v", err)
	}

	var wg sync.WaitGroup
	errorChan := make(chan error, len(depListings))

	for _, depListing := range depListings {
		wg.Add(1)
		go func(listing SimplePackageListing) {
			defer wg.Done()

			globalMu.RLock()
			alreadyDownloaded := downloadedMods[listing.Name]
			globalMu.RUnlock()

			if alreadyDownloaded {
				return
			}

			_, err := d.Download(listing)
			if err != nil {
				log.Errorf("failed to download dependency %s: %v", listing.Name, err)
				errorChan <- fmt.Errorf("failed to download dependency %s: %v", listing.Name, err)
			}
		}(depListing)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

func (d *DataService) filterDependencies(deps []string) []string {
	var filteredDeps []string
	for _, dep := range deps {
		if !strings.HasPrefix(dep, "BepInEx-BepInExPack") {
			filteredDeps = append(filteredDeps, dep)
		}
	}
	return filteredDeps
}

func (d *DataService) getModListingsForDependencies(deps []string) ([]SimplePackageListing, error) {
	var listings []SimplePackageListing
	var mu sync.Mutex
	var wg sync.WaitGroup
	errorChan := make(chan error, len(deps))

	for _, dep := range deps {
		wg.Add(1)
		go func(dependency string) {
			defer wg.Done()
			listing, err := d.getModListingForDependency(dependency)
			if err != nil {
				log.Errorf("failed to get listing for dependency %s: %v", dependency, err)
				errorChan <- fmt.Errorf("failed to get listing for dependency %s: %v", dependency, err)
				return
			}
			mu.Lock()
			listings = append(listings, listing)
			mu.Unlock()
		}(dep)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	return listings, nil
}

func (d *DataService) getModListingForDependency(dep string) (SimplePackageListing, error) {
	parts := strings.Split(dep, "-")
	if len(parts) < 2 {
		log.Errorf("invalid dependency format: %s", dep)
		return SimplePackageListing{}, fmt.Errorf("invalid dependency format: %s", dep)
	}

	version := parts[len(parts)-1]
	modName := strings.Join(parts[:len(parts)-1], "-")

	// Remove the author part from the modName for comparison
	modNameWithoutAuthor := strings.Join(strings.Split(modName, "-")[1:], "-")

	globalMu.RLock()
	defer globalMu.RUnlock()

	var matchingListing api.PackageListing
	var found bool

	for _, listing := range packageMap {
		// Compare the mod name without the author part
		if strings.EqualFold(listing.Name, modNameWithoutAuthor) {
			matchingListing = listing
			found = true
			break
		}
	}

	if !found {
		log.Errorf("mod not found: %s", modName)
		return SimplePackageListing{}, fmt.Errorf("mod not found: %s", modName)
	}

	// Find the requested version or use the latest
	var targetVersion api.Version
	for _, v := range matchingListing.Versions {
		if v.VersionNumber == version {
			targetVersion = v
			break
		}
	}
	if targetVersion.VersionNumber == "" {
		// If the specific version is not found, use the latest version
		targetVersion = matchingListing.Versions[0]
	}

	return SimplePackageListing{
		Name:        matchingListing.Name,
		Version:     targetVersion.VersionNumber,
		Description: targetVersion.Description,
		URL:         matchingListing.PackageURL,
		DownloadURL: targetVersion.DownloadURL,
		Icon:        targetVersion.Icon,
	}, nil
}

// InitializePackageMap should be called once when initializing the DataService
// func (d *DataService) InitializePackageMap() {
// 	packageMap = make(map[string]api.PackageListing, len(packageListings))
// 	for _, listing := range packageListings {
// 		packageMap[listing.Name] = listing
// 	}
// 	downloadedMods = make(map[string]bool)
// }

func (d *DataService) IsBepInExInstalled(profile profiles.Profile) bool {
	bepInExDir := filepath.Join(profile.Path, "BepInEx")
	_, err := os.Stat(bepInExDir)
	return !os.IsNotExist(err)
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

func (d *DataService) CreateProfile(name string) {
	// Check if a profile with the same name already exists
	for _, profile := range profiles.Profiles {
		if profile.Name == name {
			return
		}
	}

	// Create a new profile
	newProfilePath := filepath.Join(profiles.ProfilesDir, name)
	err = os.Mkdir(newProfilePath, os.ModePerm)
	if err != nil {
		log.Error(err)
		return
	}

	// Check if second entry in packageListings is BepInEx and install it into game folder
	if packageListings[1].Name == "BepInExPack" {

		profiles.InstallBepInEx(
			packageListings[1].Versions[0].DownloadURL,
			packageListings[1].Name,
			packageListings[1].Versions[0].VersionNumber,
			newProfilePath,
		)

		app.Events.Emit(&application.WailsEvent{
			Name: "bepinexInstalled",
			Data: true,
		})
	}

	app.Events.Emit(&application.WailsEvent{
		Name: "createdProfile",
		Data: true,
	})
}

func (d *DataService) GetProfiles() []profiles.Profile {
	profiles.Profiles = []profiles.Profile{}

	profiles.GetLocalData()

	return profiles.Profiles
}

// Windows only
// gonna have to make a system check for this, when linux support is going to come
func (d *DataService) OpenProfileDirectory(profile profiles.Profile) {
	path := filepath.Clean(profile.Path)
	cmd := exec.Command("explorer", path)
	cmd.Start()
}

func (d *DataService) SelectProfile(profile profiles.Profile) {
	app.Events.Emit(&application.WailsEvent{
		Name: "selectedProfile",
		Data: profile,
	})

	log.Info("Selected profile: ", "struct", profile)

	SelectProfile = profile

	if !d.IsBepInExInstalled(profile) {
		log.Warn("BepInEx not installed for selected profile")

		// Install BepInEx
		// should never trigger but who knows
		profiles.InstallBepInEx(
			packageListings[1].Versions[0].DownloadURL,
			packageListings[1].Name,
			packageListings[1].Versions[0].VersionNumber,
			profile.Path,
		)
	}
}
