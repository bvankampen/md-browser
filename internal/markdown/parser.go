package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)
}

// Convert parses markdown bytes and writes the parsed HTML representation to a byte slice.
func Convert(content []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
