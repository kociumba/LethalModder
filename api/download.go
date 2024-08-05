package api

import (
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
)

func Download(url, name string) (os.File, error) {
	tempFile, err := os.CreateTemp("", name)
	if err != nil {
		return os.File{}, err
	}
	// defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Failed to download file: " + resp.Status)
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		panic(err)
	}

	log.Debug(func() string { return tempFile.Name() }())

	return *tempFile, nil
}
