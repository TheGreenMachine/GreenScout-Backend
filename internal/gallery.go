// Provides the endpoint for the frontend Router Dungeon Survivor photo gallery
package internal

import (
	"os"
	"path/filepath"
)

/// Gallery Images are not to be uploaded to github.
/// Contact Tag or George for good ryan images. This is a must.

// Returns the filepath to nth image from the gallery folder.
// If one is not found, it will return an empty string, because I couldn't think of a good default image.
func GetImage(index int) string {
	allFiles, readErr := os.ReadDir(CachedConfigs.GalleryDirectory)
	if readErr != nil {
		LogError(readErr, "Unable to read gallery folder!")
	}

	for i, file := range allFiles {
		if i == index {
			return filepath.Join(CachedConfigs.GalleryDirectory, file.Name())
		}
	}
	return ""
}
