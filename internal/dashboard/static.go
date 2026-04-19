package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed templates/index.html
var templateFS embed.FS

// DashboardHandler returns an http.Handler that serves the embedded dashboard UI.
func DashboardHandler() http.Handler {
	sub, _ := fs.Sub(templateFS, "templates")
	return http.FileServer(http.FS(sub))
}
