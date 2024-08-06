//go:build !fastjson

// Better for compatibility but a bit slower
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// PackageListing represents the structure of each package listing
type PackageListing struct {
	Name           string      `json:"name"`
	FullName       string      `json:"full_name"`
	Owner          string      `json:"owner"`
	PackageURL     string      `json:"package_url"`
	DonationLink   string      `json:"donation_link"`
	DateCreated    string      `json:"date_created"`
	DateUpdated    string      `json:"date_updated"`
	UUID4          string      `json:"uuid4"`
	RatingScore    interface{} `json:"rating_score"`
	IsPinned       interface{} `json:"is_pinned"`
	IsDeprecated   interface{} `json:"is_deprecated"`
	HasNSFWContent bool        `json:"has_nsfw_content"`
	Categories     interface{} `json:"categories"`
	Versions       []Version   `json:"versions"`
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
		return err
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		log.Printf("Error unmarshaling PackageListing: %v. Raw JSON: %s", err, string(data))
		return err
	}
	return nil
}

func (p PackageListing) GetRatingScore() (int, error) {
	switch v := p.RatingScore.(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unexpected type for RatingScore: %T", v)
	}
}

func (p PackageListing) GetIsPinned() (bool, error) {
	switch v := p.IsPinned.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("unexpected type for IsPinned: %T", v)
	}
}

func (p PackageListing) GetIsDeprecated() (bool, error) {
	switch v := p.IsDeprecated.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("unexpected type for IsDeprecated: %T", v)
	}
}

func (p PackageListing) GetCategories() ([]string, error) {
	switch v := p.Categories.(type) {
	case []string:
		return v, nil
	case string:
		var categories []string
		err := json.Unmarshal([]byte(v), &categories)
		return categories, err
	default:
		return nil, fmt.Errorf("unexpected type for Categories: %T", v)
	}
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
