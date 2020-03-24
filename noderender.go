// Allow more custom node handling. Custom node can also access the default node
// handler.

package html2text

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func FromStringWithRenderer(input string, n NodeRenderer) (string, error) {
	text, err := FromReaderWithRenderer(strings.NewReader(input), n)
	if err != nil {
		return "", err
	}
	return text, nil
}

func FromReaderWithRenderer(reader io.Reader, n NodeRenderer) (string, error) {
	doc, err := html.Parse(reader)
	if err != nil {
		return "", err
	}
	return fromHtmlNodeBase(doc, n)
}

func LoopChildren(node *html.Node, d DefaultRenderer) error {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := d.DefaultRender(c); err != nil {
			return err
		}
	}
	return nil
}

// Base on 'node' and convert to plain text by writing to 'd'. 'd' can also be
// used to handle html node using the default method. Return non-nil 'err' if
// any error. Return non-nil 'next' if there is any child node that need to be
// handled by default renderer else return nil 'next' to indicate done for this
// 'node'.
type NodeRenderer interface {
	NodeRender(node *html.Node, d DefaultRendererStringWriter) (next *html.Node, err error)
}

type DefaultRendererStringWriter interface {
	DefaultRenderer
	io.StringWriter
}

type DefaultRenderer interface {
	DefaultRender(node *html.Node) error
}
