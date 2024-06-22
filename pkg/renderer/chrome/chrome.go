package chrome

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/nulltrope/thermanote/pkg/renderer"

	"context"

	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"github.com/sirupsen/logrus"
)

const (
	basePPI = 96
	// Small buffer to ensure contents fit on page
	autoPageSizeBuffer = 0.25

	defaultPixelRatio = 1
	defaultPageWidth  = 8.5
	defaultPageHeight = 11
)

var log = logrus.WithField("pkg", "renderer/chrome")

func calculateViewportSize(pageWidth, pageHeight, pixelRatio float64) (int, int) {
	scaledPPI := basePPI * pixelRatio
	return int(math.Round(pageWidth * scaledPPI)), int(math.Round(pageHeight * scaledPPI))
}

func calculatePageSize(widthPx int, heightPx int, pixelRatio float64) (float64, float64) {
	scaledPPI := basePPI * pixelRatio
	return float64(widthPx) / scaledPPI, float64(heightPx) / scaledPPI
}

type Config struct {
	viewportWidthPx  int
	viewportHeightPx int
	pageWidthIn      float64
	pageHeightIn     float64
	pixelRatio       float64
}

type Chrome struct {
	ctx context.Context

	cfg *Config
}

type ChromeRendererOption = func(c *Config)

func ViewportWidth(pixels int) ChromeRendererOption {
	return func(c *Config) {
		c.viewportWidthPx = pixels
	}
}

func ViewportHeight(pixels int) ChromeRendererOption {
	return func(c *Config) {
		c.viewportHeightPx = pixels
	}
}

func PageWidth(inches float64) ChromeRendererOption {
	return func(c *Config) {
		c.pageWidthIn = inches
	}
}

func PageHeight(inches float64) ChromeRendererOption {
	return func(c *Config) {
		c.pageHeightIn = inches
	}
}

func PixelRatio(scale float64) ChromeRendererOption {
	return func(c *Config) {
		c.pixelRatio = scale
	}
}

func NewChromeRenderer(opts ...ChromeRendererOption) (*Chrome, error) {
	ctx, _ := chromedp.NewContext(context.Background())

	c := &Chrome{
		ctx: ctx,
		cfg: &Config{
			// Set defaults that may be overridden
			pageWidthIn:  defaultPageWidth,
			pageHeightIn: defaultPageHeight,
			pixelRatio:   defaultPixelRatio,
		},
	}

	for _, opt := range opts {
		opt(c.cfg)
	}

	// Calculate the viewport
	c.calculateViewport()

	log.WithFields(logrus.Fields{
		"cfg": fmt.Sprintf("%+v", c.cfg),
	}).Debug("Created renderer")

	// Pre-allocate the browser
	return c, c.resetBrowser()
}

func (r *Chrome) calculateViewport() {
	width, height := calculateViewportSize(r.cfg.pageWidthIn, r.cfg.pageHeightIn, r.cfg.pixelRatio)
	if r.cfg.viewportWidthPx < 1 {
		r.cfg.viewportWidthPx = width
	}
	if r.cfg.viewportHeightPx < 1 {
		r.cfg.viewportHeightPx = height
	}
}

func (r *Chrome) resetBrowser() error {
	return chromedp.Run(r.ctx,
		chromedp.EmulateViewport(int64(r.cfg.viewportWidthPx), int64(r.cfg.viewportHeightPx)),
		chromedp.Navigate("about:blank"),
	)
}

func (r *Chrome) setContent(content string) error {
	return chromedp.Run(r.ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, content).Do(ctx)
		}),
	)
}

func (r *Chrome) createScreenshot() ([]byte, error) {
	imgBuf := []byte{}
	err := chromedp.Run(r.ctx,
		chromedp.Screenshot(`#text-area`, &imgBuf),
	)
	return imgBuf, err
}

func (r *Chrome) calculatePageSize(styles []*css.ComputedStyleProperty) (float64, float64) {
	pageWidth, pageHeight := r.cfg.pageWidthIn, r.cfg.pageHeightIn
	// Only iterate through style if we need to
	if r.cfg.pageWidthIn < 0 || r.cfg.pageHeightIn < 0 {
		scaledPPI := basePPI * r.cfg.pixelRatio
		for _, style := range styles {
			if style.Name == "width" && r.cfg.pageWidthIn < 0 {
				fmt.Printf("style width: %s\n", style.Value)
				if compWidth, err := strconv.ParseFloat(strings.TrimSuffix(style.Value, "px"), 64); err == nil {
					pageWidth = (float64(compWidth) / scaledPPI) + autoPageSizeBuffer
				}
			}
			if style.Name == "height" && r.cfg.pageHeightIn < 0 {
				fmt.Printf("style height: %s\n", style.Value)
				if compHeight, err := strconv.ParseFloat(strings.TrimSuffix(style.Value, "px"), 64); err == nil {
					pageHeight = (float64(compHeight) / scaledPPI) + autoPageSizeBuffer
				}
			}
		}
	}

	return pageWidth, pageHeight
}

func (r *Chrome) createPDF() ([]byte, error) {
	pdfBuf := []byte{}
	var styles []*css.ComputedStyleProperty
	err := chromedp.Run(r.ctx,
		chromedp.ComputedStyle(`#text-area`, &styles, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			width, height := r.calculatePageSize(styles)
			var err error
			log.WithFields(logrus.Fields{
				"width":  width,
				"height": height,
			}).Debugln("Creating PDF")
			pdfBuf, _, err = page.PrintToPDF().
				WithPaperWidth(width).
				WithPaperHeight(height).
				WithPrintBackground(false).
				Do(ctx)
			return err
		}),
	)
	return pdfBuf, err
}

func (r *Chrome) RenderPreview(data []byte) (*renderer.Render, error) {
	defer r.resetBrowser()

	if err := r.setContent(string(data)); err != nil {
		return nil, err
	}

	img, err := r.createScreenshot()
	if err != nil {
		return nil, err
	}

	return &renderer.Render{
		MimeType: "image/png",
		Data:     img,
	}, nil
}

func (r *Chrome) RenderPrint(data []byte) (*renderer.Render, error) {
	defer r.resetBrowser()

	if err := r.setContent(string(data)); err != nil {
		return nil, err
	}

	pdf, err := r.createPDF()
	if err != nil {
		return nil, err
	}

	return &renderer.Render{
		MimeType: "application/pdf",
		Data:     pdf,
	}, nil
}
