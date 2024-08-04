//go:build fastjson

// A bit faster but might not work
package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/valyala/fastjson"
)

// PackageListing represents the structure of each package listing
type PackageListing struct {
	Name           string
	FullName       string
	Owner          string
	PackageURL     string
	DonationLink   string
	DateCreated    string
	DateUpdated    string
	UUID4          string
	RatingScore    interface{}
	IsPinned       interface{}
	IsDeprecated   interface{}
	HasNSFWContent bool
	Categories     interface{}
	Versions       interface{}
}

// UnmarshalJSON custom unmarshaler using fastjson
func (p *PackageListing) UnmarshalJSON(data []byte) error {
	var parser fastjson.Parser
	v, err := parser.ParseBytes(data)
	if err != nil {
		return err
	}

	p.Name = string(v.GetStringBytes("name"))
	p.FullName = string(v.GetStringBytes("full_name"))
	p.Owner = string(v.GetStringBytes("owner"))
	p.PackageURL = string(v.GetStringBytes("package_url"))
	p.DonationLink = string(v.GetStringBytes("donation_link"))
	p.DateCreated = string(v.GetStringBytes("date_created"))
	p.DateUpdated = string(v.GetStringBytes("date_updated"))
	p.UUID4 = string(v.GetStringBytes("uuid4"))
	p.RatingScore = v.Get("rating_score")
	p.IsPinned = v.Get("is_pinned")
	p.IsDeprecated = v.Get("is_deprecated")
	p.HasNSFWContent = v.GetBool("has_nsfw_content")
	p.Categories = v.Get("categories")
	p.Versions = v.Get("versions")

	return nil
}

func (p PackageListing) GetRatingScore() (int, error) {
	switch v := p.RatingScore.(type) {
	case *fastjson.Value:
		if v.Type() == fastjson.TypeNumber {
			return v.GetInt(), nil
		} else if v.Type() == fastjson.TypeString {
			return strconv.Atoi(string(v.GetStringBytes()))
		}
	}
	return 0, fmt.Errorf("unexpected type for RatingScore")
}

func (p PackageListing) GetIsPinned() (bool, error) {
	switch v := p.IsPinned.(type) {
	case *fastjson.Value:
		if v.Type() == fastjson.TypeTrue {
			return true, nil
		} else if v.Type() == fastjson.TypeFalse {
			return false, nil
		} else if v.Type() == fastjson.TypeString {
			return strconv.ParseBool(string(v.GetStringBytes()))
		}
	}
	return false, fmt.Errorf("unexpected type for IsPinned")
}

func (p PackageListing) GetIsDeprecated() (bool, error) {
	switch v := p.IsDeprecated.(type) {
	case *fastjson.Value:
		if v.Type() == fastjson.TypeTrue {
			return true, nil
		} else if v.Type() == fastjson.TypeFalse {
			return false, nil
		} else if v.Type() == fastjson.TypeString {
			return strconv.ParseBool(string(v.GetStringBytes()))
		}
	}
	return false, fmt.Errorf("unexpected type for IsDeprecated")
}

func (p PackageListing) GetCategories() ([]string, error) {
	switch v := p.Categories.(type) {
	case *fastjson.Value:
		if v.Type() == fastjson.TypeArray {
			var categories []string
			for _, cat := range v.GetArray() {
				categories = append(categories, string(cat.GetStringBytes()))
			}
			return categories, nil
		} else if v.Type() == fastjson.TypeString {
			var categories []string
			err := json.Unmarshal(v.GetStringBytes(), &categories)
			return categories, err
		}
	}
	return nil, fmt.Errorf("unexpected type for Categories")
}

func GetMods() ([]PackageListing, error) {
	// resp, err := http.Get("https://thunderstore.io/api/v1/package/")
	resp, err := http.Get("https://thunderstore.io/c/lethal-company/api/v1/package/")
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var parser fastjson.Parser
	v, err := parser.ParseBytes(body)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	var packageListings []PackageListing
	for _, item := range v.GetArray() {
		var pl PackageListing
		err := pl.UnmarshalJSON(item.MarshalTo(nil))
		if err != nil {
			log.Printf("Error unmarshaling package: %v", err)
		} else {
			packageListings = append(packageListings, pl)
		}
	}

	return packageListings, nil
}
