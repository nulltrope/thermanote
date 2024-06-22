package renderer

type Render struct {
	MimeType string
	Data     []byte
}

type Renderer interface {
	RenderPreview([]byte) (*Render, error)
	RenderPrint([]byte) (*Render, error)
}
