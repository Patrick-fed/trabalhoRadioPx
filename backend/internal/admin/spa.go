package admin

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web/*
var webFS embed.FS

// SPAHandler serves the embedded admin UI at /admin and /admin/*.
// Static assets come from web/; unknown subpaths fall back to index.html
// (single-page app).
type SPAHandler struct {
	files http.Handler
	root  fs.FS
}

// NewSPAHandler builds the /admin file server from the embedded web/ tree.
func NewSPAHandler() (*SPAHandler, error) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil, err
	}
	return &SPAHandler{
		files: http.FileServer(http.FS(sub)),
		root:  sub,
	}, nil
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Paths arrive as /admin or /admin/<asset>; the FileServer expects
	// paths relative to web/. Serve the directory root for the index so
	// FileServer does not redirect /index.html → ./ (which loops under
	// the /admin prefix).
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "" || path == "/" {
		path = "/"
	}

	name := strings.TrimPrefix(path, "/")
	if name != "" && name != "/" {
		if _, err := fs.Stat(h.root, name); err != nil {
			// Unknown subpath → SPA fallback (directory root serves index.html).
			path = "/"
		}
	}

	r2 := r.Clone(r.Context())
	r2.URL.Path = path
	h.files.ServeHTTP(w, r2)
}
