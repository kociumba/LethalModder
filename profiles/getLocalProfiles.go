package profiles

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var (
	UserConfigDir string
	DataDir       string
	ProfilesDir   string
	Profiles      []Profile

	err error
)

type Profile struct {
	Name string   `json:"name"`
	Path string   `json:"path"`
	Mods []string `json:"mods"`
}

func GetLocalData() {
	UserConfigDir, err = os.UserConfigDir()
	if err != nil {
		log.Error(err)
		return
	}

	DataDir = filepath.Join(UserConfigDir, "LethalModder")

	_, err = os.Stat(DataDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(DataDir, os.ModePerm)
		if err != nil {
			log.Error(err)
			return
		}
	}

	ProfilesDir = filepath.Join(DataDir, "profiles")

	// I honestly don't know why I'm checking twice and not just once with MkdirAll
	_, err = os.Stat(ProfilesDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(ProfilesDir, os.ModePerm)
		if err != nil {
			log.Error(err)
			return
		}
	}

	// Iterate over every directory in profiles/
	entries, err := os.ReadDir(ProfilesDir)
	if err != nil {
		log.Error(err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			profileName := entry.Name()
			profilePath := filepath.Join(ProfilesDir, profileName)

			mods, err := loadMods(profilePath)
			if err != nil {
				log.Error(err)
				continue
			}

			profile := Profile{
				Name: profileName,
				Path: profilePath,
				Mods: mods,
			}
			Profiles = append(Profiles, profile)
		}
	}
}

// loadMods loads the mods from a given profile directory.
func loadMods(profilePath string) ([]string, error) {
	modsPath := filepath.Join(profilePath, "BepInEx")
	modFiles, err := os.ReadDir(modsPath)
	if err != nil {
		return nil, err
	}

	var mods []string
	for _, mod := range modFiles {
		if mod.IsDir() {
			mods = append(mods, mod.Name())
		}
	}

	return mods, nil
}
