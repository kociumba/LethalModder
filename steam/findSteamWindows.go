//go:build !linux

package steam

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/andygrunwald/vdf"
	"github.com/charmbracelet/log"
	"golang.org/x/sys/windows/registry"
)

func FindSteam() (steamPath string, lethalCompanyPath string, err error) {
	// Get the Steam installation path from the registry
	steamPath, err = getSteamInstallPathFromRegistry()
	if err != nil {
		return "", "", err
	}

	// Find the installation path of "Lethal Company"
	lethalCompanyPath, err = findGameInstallPath(steamPath, name)
	if err != nil {
		return steamPath, "", err
	}

	// Test if the path contains "Lethal Company.exe"
	if _, err := os.Stat(filepath.Join(lethalCompanyPath, "Lethal Company.exe")); err == nil {
		return steamPath, lethalCompanyPath, nil
	}

	return steamPath, lethalCompanyPath, nil
}

func getSteamInstallPathFromRegistry() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	steamPath, _, err := key.GetStringValue("InstallPath")
	if err != nil {
		return "", err
	}

	return steamPath, nil
}

func findGameInstallPath(steamPath string, gameName string) (string, error) {
	libraryFilePath := filepath.Join(steamPath, "steamapps", "libraryfolders.vdf")
	file, err := os.Open(libraryFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Parse the VDF file
	p := vdf.NewParser(file)
	data, err := p.Parse()
	if err != nil {
		return "", err
	}

	libraries, ok := data["libraryfolders"].(map[string]interface{})
	if !ok {
		return "", errVDFParse
	}

	// log.Debug("Parsed content", "libraries", libraries)

	// Check each library for the game
	for _, v := range libraries {
		steamLibrary, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		log.Debug("Checking folder", "path", steamLibrary["path"])

		// Check for the game in this library
		lethalCompanyPath, err := checkForGame(steamLibrary["path"].(string), gameName)
		if err == nil {
			return lethalCompanyPath, nil // Return the found path immediately
		}
	}

	return "", errNoGame // Return an error if the game is not found in any library
}

// I'm a fucking idiot for this
// I could have just went through steamapps/common/ and looked if there is a Lethal Company directory
func checkForGame(path string, name string) (string, error) {
	manifestPath := filepath.Join(path, "steamapps")

	files, err := filepath.Glob(filepath.Join(manifestPath, "appmanifest_*.acf"))
	if err != nil {
		log.Errorf("Error getting appmanifest files: %v", err)
		return "", err
	}

	for _, file := range files {
		fileData, err := os.ReadFile(file)
		if err != nil {
			log.Errorf("Error reading appmanifest file: %v", err)
			continue
		}

		p := vdf.NewParser(bytes.NewReader(fileData))
		data, err := p.Parse()
		if err != nil {
			log.Errorf("Error parsing appmanifest file: %v", err)
			continue
		}

		appState, ok := data["AppState"].(map[string]interface{})
		if !ok {
			log.Errorf("Error parsing AppState in appmanifest file: %v", err)
			continue
		}

		nameStr, ok := appState["name"].(string)
		if !ok {
			log.Errorf("Error parsing name in appmanifest file: %v", err)
			continue
		}

		nameStr = strings.Trim(nameStr, "\"")

		if nameStr == name {
			file = getActualPath(path)

			log.Debug("Found game", "path", path, "file", file) // Log the matching file
			return file, nil                                    // Return the found file path
		}

		log.Debug("App not found for gameID", "name", name, "found name", nameStr, "file", file)
	}

	return "", errNoGame // Return an error if the game is not found in this library
}

func getActualPath(path string) string {
	path = filepath.Join(path, "steamapps", "common", name)

	// Test if this exists
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}
