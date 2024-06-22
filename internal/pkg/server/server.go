package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nulltrope/thermanote/pkg/printer/ipp"
	"github.com/nulltrope/thermanote/pkg/renderer/chrome"

	"github.com/sirupsen/logrus"
)

const apiPrefix = "/api/v1/"

var log = logrus.WithField("pkg", "server")

type HTTPServer struct {
	httpSrv *http.Server
}

func NewServer(address string) (*HTTPServer, error) {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix(apiPrefix).Subrouter()

	logrus.SetLevel(logrus.DebugLevel)

	chromeRenderer, err := chrome.NewChromeRenderer(
		chrome.PageWidth(4),
		// Size page to size of viewport
		chrome.PageHeight(1),
		// Really high viewport for "infinite" length
		chrome.ViewportHeight(10000),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't create chrome renderer: %v", err)
	}

	ippPrinter := ipp.NewIPPPrinter()

	// Preview Handler
	previewHandler := newPreviewHandler(chromeRenderer)
	apiRouter.Methods("POST").Path("/preview").Handler(previewHandler)

	// Print handler
	printHandler := newPrintHandler(chromeRenderer, ippPrinter)
	apiRouter.Methods("POST").Path("/print").Handler(printHandler)

	staticHandler, err := newStaticHandler()
	if err != nil {
		return nil, fmt.Errorf("error creating static handler: %v", err)
	}
	router.Methods("GET").PathPrefix("/frontend/static").Handler(staticHandler)

	// Catch-all, must be last
	router.Methods("GET").PathPrefix("/").Handler(staticHandler)

	return &HTTPServer{
		httpSrv: &http.Server{
			Handler:      router,
			Addr:         address,
			WriteTimeout: 30 * time.Second,
			ReadTimeout:  30 * time.Second,
		},
	}, nil
}

func (s *HTTPServer) Start() error {
	log.WithField("address", s.httpSrv.Addr).Info("Started HTTP server")
	return s.httpSrv.ListenAndServe()
}
