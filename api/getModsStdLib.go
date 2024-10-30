//go:build !fastjson

// Better for compatibility but a bit slower
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

// PackageListing represents the structure of each package listing
type PackageListing struct {
	Name           string    `json:"name"`
	FullName       string    `json:"full_name"`
	Owner          string    `json:"owner"`
	PackageURL     string    `json:"package_url"`
	DonationLink   string    `json:"donation_link"`
	DateCreated    string    `json:"date_created"`
	DateUpdated    string    `json:"date_updated"`
	UUID4          string    `json:"uuid4"`
	RatingScore    uint32    `json:"rating_score"`
	IsPinned       bool      `json:"is_pinned"`
	IsDeprecated   bool      `json:"is_deprecated"`
	HasNSFWContent bool      `json:"has_nsfw_content"`
	Categories     []string  `json:"categories"`
	Versions       []Version `json:"versions"` // this may have something to do with the error
}

// Version represents the structure of a version listing
type Version struct {
	DateCreated   string   `json:"date_created"`
	Dependencies  []string `json:"dependencies"`
	Description   string   `json:"description"`
	DownloadURL   string   `json:"download_url"`
	Downloads     int      `json:"downloads"`
	FileSize      int      `json:"file_size"`
	FullName      string   `json:"full_name"`
	Icon          string   `json:"icon"`
	IsActive      bool     `json:"is_active"`
	Name          string   `json:"name"`
	UUID4         string   `json:"uuid4"`
	VersionNumber string   `json:"version_number"`
	WebsiteURL    string   `json:"website_url"`
}

func (p *PackageListing) UnmarshalJSON(data []byte) error {
	type Alias PackageListing
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		log.Errorf("Error unmarshaling PackageListing: %v. Raw JSON: %s", err, string(data))
		return err
	}
	return nil
}

func GetMods() ([]PackageListing, error) {
	// resp, err := http.Get("https://thunderstore.io/api/v1/package/")
	resp, err := http.Get("https://thunderstore.io/c/lethal-company/api/v1/package/")
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var packageListings []PackageListing
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&packageListings)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v, decoded: %#v", err, packageListings)
	}

	return packageListings, nil
}

func (p PackageListing) Latest() Version {
	return p.Versions[0]
}

func (p PackageListing) IsModpack() bool {
	for _, category := range p.Categories {
		if strings.Contains(category, "Modpacks") {
			return true
		}
	}

	return false
}
