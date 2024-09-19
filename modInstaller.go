//go:build !old

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kociumba/LethalModder/api"
	"github.com/kociumba/LethalModder/profiles"
)

type InstallArgs struct {
	Mod         SimplePackageListing
	PackagePath string
	Profile     profiles.Profile
}

type Profile struct {
	Path string
}

func (p Profile) GetPathOfProfile() string {
	return p.Path
}

type ModLoaderPackage struct {
	PackageName string
	RootFolder  string
	LoaderType  string
}

var MODLOADER_PACKAGES = []ModLoaderPackage{
	{PackageName: "BepInExPack", RootFolder: "BepInEx"},
}

var basePackageFiles = []string{"manifest.json", "readme.md", "icon.png"}

// BUG: Double extraction, when downloading a mod it extracts next to BepInEx/ but also correctly in BepInEx/plugins/
func (d *DataService) Download(listing SimplePackageListing) (*string, error) {
	if SelectProfile.Path == "" {
		return nil, fmt.Errorf("no profile selected")
	}

	bepInExDir := filepath.Join(SelectProfile.Path, "BepInEx")

	// Download the mod file
	tempFile, err := downloadFile(listing.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download mod: %w", err)
	}
	defer os.Remove(tempFile)

	// Extract the mod and parse its manifest
	extractedDir, manifest, err := extractAndParseManifest(tempFile, bepInExDir, listing.Name, listing.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to extract and parse manifest: %w", err)
	}

	// Install the mod
	installer := &BepInExInstaller{}
	installArgs := InstallArgs{
		Mod:         listing,
		PackagePath: extractedDir,
		Profile:     SelectProfile,
	}
	if err := installer.Install(installArgs); err != nil {
		return nil, fmt.Errorf("failed to install mod: %w", err)
	}

	// Handle dependencies
	if err := handleDependencies(manifest, bepInExDir); err != nil {
		return nil, fmt.Errorf("failed to handle dependencies: %w", err)
	}

	return &extractedDir, nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: %s", resp.Status)
	}

	tempFile, err := os.CreateTemp("", "mod-*.zip")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func extractAndParseManifest(src, dest, modName, modVersion string) (string, *ModManifest, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", nil, err
	}
	defer r.Close()

	extractedDir := filepath.Join(dest, "plugins", fmt.Sprintf("%s_%s", modName, modVersion))
	var manifestFile *zip.File

	for _, f := range r.File {
		fpath := filepath.Join(extractedDir, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(extractedDir)+string(os.PathSeparator)) {
			return "", nil, fmt.Errorf("invalid file path")
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", nil, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", nil, err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", nil, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", nil, err
		}

		if filepath.Base(f.Name) == "manifest.json" {
			manifestFile = f
		}
	}

	if manifestFile == nil {
		return "", nil, fmt.Errorf("manifest.json not found")
	}

	manifest, err := parseManifest(manifestFile)
	if err != nil {
		return "", nil, err
	}

	return extractedDir, manifest, nil
}

func parseManifest(file *zip.File) (*ModManifest, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// Wrap the rc with the skipBOM function to handle BOM removal
	reader := skipBOM(rc)

	var manifest ModManifest
	err = json.NewDecoder(reader).Decode(&manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func skipBOM(r io.Reader) io.Reader {
	// Use a buffer to peek at the first 3 bytes to check for the BOM
	buf := make([]byte, 3)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return r // Return the original reader if we can't read
	}

	// Check for BOM (0xEF, 0xBB, 0xBF)
	if n >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		return io.MultiReader(bytes.NewReader(buf[3:n]), r) // Skip the BOM
	}

	// No BOM found, return a reader that includes the bytes we already read
	return io.MultiReader(bytes.NewReader(buf[:n]), r)
}

// BUG: deps treat BepInEx/plugins/ as root and try to find another BepInEx/ in there
//
// Need to fix this to avoid unpacking errors
func handleDependencies(manifest *ModManifest, bepInExDir string) error {
	for _, dep := range manifest.Dependencies {
		if strings.HasPrefix(dep, "BepInEx-BepInExPack") {
			continue
		}

		depListing, err := getModListingForDependency(dep)
		if err != nil {
			return err
		}

		dataService := &DataService{}
		_, err = dataService.Download(depListing)
		if err != nil {
			return err
		}
	}

	return nil
}

func getModListingForDependency(dep string) (SimplePackageListing, error) {
	parts := strings.Split(dep, "-")
	if len(parts) < 2 {
		return SimplePackageListing{}, fmt.Errorf("invalid dependency format: %s", dep)
	}

	version := parts[len(parts)-1]
	modName := strings.Join(parts[:len(parts)-1], "-")

	// Remove the author part from the modName for comparison
	modNameWithoutAuthor := strings.Join(strings.Split(modName, "-")[1:], "-")

	var matchingListing api.PackageListing
	var found bool

	for _, listing := range packageListings {
		if strings.EqualFold(listing.Name, modNameWithoutAuthor) {
			matchingListing = listing
			found = true
			break
		}
	}

	if !found {
		return SimplePackageListing{}, fmt.Errorf("mod not found: %s", modName)
	}

	return convertToSimplePackageListing(matchingListing, version), nil
}

func convertToSimplePackageListing(pl api.PackageListing, version string) SimplePackageListing {
	var targetVersion api.Version
	for _, v := range pl.Versions {
		if v.VersionNumber == version {
			targetVersion = v
			break
		}
	}

	// If the specific version is not found, use the latest version
	if targetVersion.VersionNumber == "" && len(pl.Versions) > 0 {
		targetVersion = pl.Versions[0]
	}

	return SimplePackageListing{
		Name:        pl.Name,
		Version:     targetVersion.VersionNumber,
		Description: targetVersion.Description,
		URL:         pl.PackageURL,
		DownloadURL: targetVersion.DownloadURL,
		Icon:        targetVersion.Icon,
	}
}

type BepInExInstaller struct{}

func (i *BepInExInstaller) Install(args InstallArgs) error {
	mapping := findMapping(args.Mod.Name)
	mappingRoot := ""
	if mapping != nil {
		mappingRoot = mapping.RootFolder
	}

	var bepInExRoot string
	if mappingRoot != "" {
		bepInExRoot = filepath.Join(args.PackagePath, mappingRoot)
	} else {
		bepInExRoot = args.PackagePath
	}

	items, err := os.ReadDir(bepInExRoot)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, item := range items {
		if !contains(basePackageFiles, strings.ToLower(item.Name())) {
			srcPath := filepath.Join(bepInExRoot, item.Name())
			destPath := filepath.Join(args.Profile.Path, item.Name())

			info, err := os.Stat(srcPath)
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}

			if info.IsDir() {
				if err := copyFolder(srcPath, destPath); err != nil {
					return fmt.Errorf("failed to copy folder: %w", err)
				}
			} else {
				if err := copyFile(srcPath, destPath); err != nil {
					return fmt.Errorf("failed to copy file: %w", err)
				}
			}
		}
	}

	return nil
}

func findMapping(modName string) *ModLoaderPackage {
	for _, entry := range MODLOADER_PACKAGES {
		if strings.EqualFold(entry.PackageName, modName) && entry.LoaderType == "BEPINEX" {
			return &entry
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func copyFile(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dest, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

func copyFolder(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyFolder(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := copySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copyFile(sourcePath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, perm); err != nil {
			return fmt.Errorf("failed to create directory: %s, error: %w", dir, err)
		}
	}
	return nil
}

func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
