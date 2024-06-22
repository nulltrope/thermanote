package server

import (
	"bytes"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type templateConfig struct {
	PreviewEndpoint string
	PrintEndpoint   string
}

type staticHandler struct {
	template *template.Template
}

func newStaticHandler() (staticHandler, error) {
	tmpl, err := template.ParseGlob("frontend/templates/*")
	if err != nil {
		return staticHandler{}, fmt.Errorf("error parsing template glob: %v", err)
	}

	return staticHandler{
		template: tmpl,
	}, nil
}

func (h staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cleanedPath := filepath.Clean(r.URL.Path)
	if cleanedPath == "/" {
		cleanedPath = "index.html"
	}
	cleanedPath = strings.TrimPrefix(cleanedPath, "/")

	var err error
	var data = make([]byte, 0)
	// First see if this file maps to a template
	template := h.template.Lookup(cleanedPath)
	if template != nil {
		var buf bytes.Buffer
		err = template.Execute(&buf, templateConfig{
			PreviewEndpoint: "/api/v1/preview",
			PrintEndpoint:   "/api/v1/print",
		})
		data = buf.Bytes()
	} else {
		data, err = os.ReadFile(fmt.Sprintf("frontend/static/%s", cleanedPath))
	}

	if err != nil {
		fmt.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Bad request")
		return
	}

	ext := mime.TypeByExtension(filepath.Ext(cleanedPath))
	w.Header().Add("Content-Type", ext)
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(data))
}
