package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/nulltrope/thermanote/pkg/renderer"
)

type previewHandler struct {
	renderer renderer.Renderer
}

func newPreviewHandler(renderer renderer.Renderer) previewHandler {
	return previewHandler{
		renderer: renderer,
	}
}

func (h previewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log = log.WithField("handler", "preview")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Bad request")
		return
	}
	defer r.Body.Close()
	log.WithField("bytes", len(bodyBytes)).Debug("Received bytes")

	img, err := h.renderer.RenderPreview(bodyBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error generating preview: %v", err)
		return
	}
	log.
		WithField("bytes", len(img.Data)).
		WithField("mimeType", img.MimeType).
		Debug("Generated preview")

	w.Header().Add("Content-Type", img.MimeType)
	w.Header().Add("Content-Transfer-Encoding", "binary")
	_, err = w.Write(img.Data)
	if err != nil {
		log.Errorf("Error writing preview data: %v", err)
	}
}
