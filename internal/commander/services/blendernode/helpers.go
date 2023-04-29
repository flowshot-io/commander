package blendernode

import (
	"path/filepath"
	"regexp"
	"strings"
)

func validateArtifactName(name string) bool {
	// Check if the name contains any invalid characters
	if !regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`).MatchString(name) {
		return false
	}

	// Check if the name contains any file extension
	ext := filepath.Ext(name)
	return ext == ""
}

func getArtifactFileName(name string) string {
	return name + ".tar.gz"
}

func isArtifactFile(file string) bool {
	// Check if the file has a .tar.gz extension
	return strings.HasSuffix(file, ".tar.gz")
}
