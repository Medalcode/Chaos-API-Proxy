package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var content embed.FS

// Handler returns a http.Handler that serves the UI
// Since we are not building a dist folder, we will serve from the web folder in dev mode
// or embed it. For simplicity in this project structure without a build step:
// We will point to the ../../web folder relative to execution or embed it if we moved it.

// Let's assume we want to embed the 'web' folder content.
// NOTE: Go embed directive is relative to the package directory.
// We need to put the html file where it can be embedded or use os.DirFS for development.

// Better approach for this structure: Let's create a AssetsHandler that serves from filesystem for now
// to avoid complex moving of files.
func Handler() http.Handler {
	return http.FileServer(http.Dir("./web"))
}
