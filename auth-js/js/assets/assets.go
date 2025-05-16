package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Manifest struct {
	JS  string `json:"js"`
	CSS string `json:"css"`
}

var manifest *Manifest

// LoadManifest loads the asset manifest from the static directory
func LoadManifest(staticDir string) error {
	manifestPath := filepath.Join(staticDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	manifest = &Manifest{}
	return json.Unmarshal(data, manifest)
}

// GetJS returns the path to the JavaScript file
func GetJS() string {
	if manifest == nil {
		return "/static/main.js" // fallback
	}
	return "/static/" + manifest.JS
}

// GetCSS returns the path to the CSS file
func GetCSS() string {
	if manifest == nil {
		return "/static/main.css" // fallback
	}
	return "/static/" + manifest.CSS
}
