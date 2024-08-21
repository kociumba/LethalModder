package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func (d *DataService) IsBepInExInstalled(profile profiles.Profile) bool {
	bepInExDir := filepath.Join(profile.Path, "BepInEx")
	_, err := os.Stat(bepInExDir)
	return !os.IsNotExist(err)
}
