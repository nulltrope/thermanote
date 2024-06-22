package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nulltrope/thermanote/pkg/printer"
	"github.com/nulltrope/thermanote/pkg/renderer"
)

type printHandler struct {
	renderer renderer.Renderer
	printer  printer.Printer
}

func newPrintHandler(renderer renderer.Renderer, printer printer.Printer) printHandler {
	return printHandler{
		renderer: renderer,
		printer:  printer,
	}
}

func (h printHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log = log.WithField("handler", "print")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Bad request")
		return
	}
	defer r.Body.Close()
	log.WithField("bytes", len(bodyBytes)).Debug("Received bytes")

	pdf, err := h.renderer.RenderPrint(bodyBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error generating print: %v", err)
		return
	}
	log.
		WithField("bytes", len(pdf.Data)).
		WithField("mimeType", pdf.MimeType).
		Debug("Generated print")

	if strings.ToLower(r.URL.Query().Get("preview")) == "true" {
		w.Header().Add("Content-Type", pdf.MimeType)
		w.Header().Add("Content-Transfer-Encoding", "binary")
		_, err = w.Write(pdf.Data)
		if err != nil {
			log.Errorf("Error writing print data: %v", err)

		}
		return
	}

	err = h.printer.Print(pdf.Data)
	if err != nil {
		log.Errorf("Error printing data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error printing data: %v", err)
		return
	}
	w.Write([]byte("success"))
}
