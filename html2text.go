package html2text

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	spacingRe = regexp.MustCompile(`[ \r\n\t]+`)
	newlineRe = regexp.MustCompile(`\n\n+`)
)

type textifyTraverseCtx struct {
	Buf bytes.Buffer

	customRender    NodeRenderer
	prefix          string
	blockquoteLevel int
	lineLength      int
	endsWithSpace   bool
	endsWithNewline bool
	justClosedDiv   bool
}

func (ctx *textifyTraverseCtx) traverse(node *html.Node) error {
	return ctx.DefaultRender(node)
}

// Implements DefaultRenderer.
func (ctx *textifyTraverseCtx) DefaultRender(node *html.Node) error {
	if ctx.customRender != nil {
		next, err := ctx.customRender.NodeRender(node, ctx)
		if err != nil {
			return err
		} else if next == nil {
			return nil
		}
		node = next
	}
	switch node.Type {

	default:
		return ctx.traverseChildren(node)

	case html.TextNode:
		data := strings.Trim(spacingRe.ReplaceAllString(node.Data, " "), " ")
		return ctx.emit(data)

	case html.ElementNode:

		ctx.justClosedDiv = false
		switch node.DataAtom {
		case atom.Br:
			return ctx.emit("\n")

		case atom.H1, atom.H2, atom.H3:
			subCtx := textifyTraverseCtx{}
			if err := subCtx.traverseChildren(node); err != nil {
				return err
			}

			str := subCtx.Buf.String()
			dividerLen := 0
			for _, line := range strings.Split(str, "\n") {
				if lineLen := len([]rune(line)); lineLen-1 > dividerLen {
					dividerLen = lineLen - 1
				}
			}
			divider := ""
			if node.DataAtom == atom.H1 {
				divider = strings.Repeat("*", dividerLen)
			} else {
				divider = strings.Repeat("-", dividerLen)
			}

			if node.DataAtom == atom.H3 {
				return ctx.emit("\n\n" + str + "\n" + divider + "\n\n")
			}
			return ctx.emit("\n\n" + divider + "\n" + str + "\n" + divider + "\n\n")

		case atom.Blockquote:
			ctx.blockquoteLevel++
			ctx.prefix = strings.Repeat(">", ctx.blockquoteLevel) + " "
			if err := ctx.emit("\n"); err != nil {
				return err
			}
			if ctx.blockquoteLevel == 1 {
				if err := ctx.emit("\n"); err != nil {
					return err
				}
			}
			if err := ctx.traverseChildren(node); err != nil {
				return err
			}
			ctx.blockquoteLevel--
			ctx.prefix = strings.Repeat(">", ctx.blockquoteLevel)
			if ctx.blockquoteLevel > 0 {
				ctx.prefix += " "
			}
			return ctx.emit("\n\n")

		case atom.Div:
			if ctx.lineLength > 0 {
				if err := ctx.emit("\n"); err != nil {
					return err
				}
			}
			if err := ctx.traverseChildren(node); err != nil {
				return err
			}
			var err error
			if ctx.justClosedDiv == false {
				err = ctx.emit("\n")
			}
			ctx.justClosedDiv = true
			return err

		case atom.Li:
			if err := ctx.emit("* "); err != nil {
				return err
			}

			if err := ctx.traverseChildren(node); err != nil {
				return err
			}

			return ctx.emit("\n")

		case atom.B, atom.Strong:
			subCtx := textifyTraverseCtx{}
			subCtx.endsWithSpace = true
			if err := subCtx.traverseChildren(node); err != nil {
				return err
			}
			str := subCtx.Buf.String()
			return ctx.emit("*" + str + "*")

		case atom.A:
			// If image is the only child, take its alt text as the link text
			if img := node.FirstChild; img != nil && node.LastChild == img && img.DataAtom == atom.Img {
				if altText := getAttrVal(img, "alt"); altText != "" {
					ctx.emit(altText)
				}
			} else if err := ctx.traverseChildren(node); err != nil {
				return err
			}

			hrefLink := ""
			if attrVal := getAttrVal(node, "href"); attrVal != "" {
				attrVal = ctx.normalizeHrefLink(attrVal)
				if attrVal != "" {
					hrefLink = "( " + attrVal + " )"
				}
			}

			return ctx.emit(hrefLink)

		case atom.P, atom.Ul, atom.Table:
			if err := ctx.emit("\n\n"); err != nil {
				return err
			}

			if err := ctx.traverseChildren(node); err != nil {
				return err
			}

			return ctx.emit("\n\n")

		case atom.Tr:
			if err := ctx.traverseChildren(node); err != nil {
				return err
			}

			return ctx.emit("\n")

		case atom.Style, atom.Script, atom.Head:
			// Ignore the subtree
			return nil

		default:
			return ctx.traverseChildren(node)
		}
	}
}

func (ctx *textifyTraverseCtx) traverseChildren(node *html.Node) error {
	return LoopChildren(node, ctx)
}

func (ctx *textifyTraverseCtx) Write(b []byte) (int, error) {
	if err := ctx.emit(string(b)); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (ctx *textifyTraverseCtx) emit(data string) error {
	if len(data) == 0 {
		return nil
	}
	lines := ctx.breakLongLines(data)
	var err error
	for _, line := range lines {
		runes := []rune(line)
		startsWithSpace := unicode.IsSpace(runes[0])
		if !startsWithSpace && !ctx.endsWithSpace {
			ctx.Buf.WriteByte(' ')
			ctx.lineLength++
		}
		ctx.endsWithSpace = unicode.IsSpace(runes[len(runes)-1])
		for _, c := range line {
			_, err = ctx.Buf.WriteString(string(c))
			if err != nil {
				return err
			}
			ctx.lineLength++
			if c == '\n' {
				ctx.lineLength = 0
				if ctx.prefix != "" {
					_, err = ctx.Buf.WriteString(ctx.prefix)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (ctx *textifyTraverseCtx) breakLongLines(data string) []string {
	// only break lines when we are in blockquotes
	if ctx.blockquoteLevel == 0 {
		return []string{data}
	}
	var ret []string
	runes := []rune(data)
	l := len(runes)
	existing := ctx.lineLength
	if existing >= 74 {
		ret = append(ret, "\n")
		existing = 0
	}
	for l+existing > 74 {
		i := 74 - existing
		for i >= 0 && !unicode.IsSpace(runes[i]) {
			i--
		}
		if i == -1 {
			// no spaces, so go the other way
			i = 74 - existing
			for i < l && !unicode.IsSpace(runes[i]) {
				i++
			}
		}
		ret = append(ret, string(runes[:i])+"\n")
		for i < l && unicode.IsSpace(runes[i]) {
			i++
		}
		runes = runes[i:]
		l = len(runes)
		existing = 0
	}
	if len(runes) > 0 {
		ret = append(ret, string(runes))
	}
	return ret
}

func (ctx *textifyTraverseCtx) normalizeHrefLink(link string) string {
	link = strings.TrimSpace(link)
	link = strings.TrimPrefix(link, "mailto:")
	return link
}

func getAttrVal(node *html.Node, attrName string) string {
	v, _ := getAttrValExist(node, attrName)
	return v
}

func getAttrValExist(node *html.Node, attrName string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val, true
		}
	}
	return "", false
}

// FromHTMLNode renders text output from a pre-parsed HTML document.
func FromHtmlNode(doc *html.Node) (string, error) {
	return fromHtmlNodeBase(doc, nil)
}

// FromReader renders text output after parsing HTML for the specified
// io.Reader.
func FromReader(reader io.Reader) (string, error) {
	return FromReaderWithRenderer(reader, nil)
}

// FromString parses HTML from the input string, then renders the text form.
func FromString(input string) (string, error) {
	return FromStringWithRenderer(input, nil)
}

func fromHtmlNodeBase(doc *html.Node, n NodeRenderer) (string, error) {
	ctx := textifyTraverseCtx{
		Buf:          bytes.Buffer{},
		customRender: n,
	}
	if err := ctx.traverse(doc); err != nil {
		return "", err
	}

	text := strings.TrimSpace(newlineRe.ReplaceAllString(
		strings.Replace(ctx.Buf.String(), "\n ", "\n", -1), "\n\n"))
	return text, nil

}
