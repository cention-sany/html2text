package html2text

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestErrorHandling(t *testing.T) {
	er := errorer{}
	if _, err := FromReader(er); err != errTest {
		assertError(t, errTest, err)
	}
	if _, err := FromReaderWithRenderer(er, nil); err != errTest {
		assertError(t, errTest, err)
	}
	const (
		nodeInput   = "<p>ERROR</p>"
		plainOutput = "ERROR"
	)
	node, err := html.Parse(strings.NewReader(nodeInput))
	if err != nil {
		t.Fatal(err)
	}
	output, err := FromHtmlNode(node)
	if err != nil {
		t.Fatal(err)
	} else if output != plainOutput {
		t.Errorf("Expect output=%s but got=%s\n", plainOutput, output)
	}
	if err = LoopChildren(node, er); err != errTest {
		assertError(t, errTest, err)
	}
	_, err = FromReaderWithRenderer(strings.NewReader(nodeInput), er)
	if err != errTest {
		assertError(t, errTest, err)
	}
	_, err = FromStringWithRenderer(nodeInput, er)
	if err != errTest {
		assertError(t, errTest, err)
	}
}

func assertError(t *testing.T, expected, result error) {
	t.Errorf("Expect error=%v but got=%v\n", expected, result)
}

type errorer struct{}

func (errorer) Read(_ []byte) (int, error) {
	return 0, errTest
}

func (errorer) DefaultRender(_ *html.Node) error {
	return errTest
}

func (errorer) NodeRender(_ *html.Node, _ DefaultRendererWriter) (*html.Node, error) {
	return nil, errTest
}

var errTest = errors.New("test error")

func TestCustomRenderer(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"Test text span",
			"Test text span",
		},
		{
			"Test text span<br>",
			"Test text span",
		},
		{
			"Test text span<br>Test",
			"Test text span\nTest",
		},
		{
			"<p>P1</p><p>P2</p>",
			"P1\n\nP2",
		},
		{
			`<p>P1</p><p><span data-custom alt="Special text start"></span>P2</p><span data-custom alt="Special text end"></span>`,
			"P1\n\nSpecial text start P2\n\nSpecial text end",
		},
		{
			`<p>P1</p><p><span alt="Special text start"></span>P2</p><span alt="Special text end"></span>`,
			"P1\n\nP2",
		},
		{
			`<p>P1</p><p><span data-custom alt="START"><b>S1</b></span>P2</p><span data-custom alt="END"><b>S2</b></span>`,
			"P1\n\nSTART *S1* P2\n\nEND *S2*",
		},
		{
			`<p>P1</p><p><span>S1</span>P2</p><span>S2</span>`,
			"P1\n\nS1 P2\n\nS2",
		},
	}

	for _, testCase := range testCases {
		assertStringWithRenderer(t, testCase.input, testCase.output,
			altSpanText{})
	}
}

type altSpanText struct{}

func (altSpanText) NodeRender(node *html.Node, d DefaultRendererWriter) (*html.Node, error) {
	if node.Type != html.ElementNode {
		return node, nil
	} else if node.DataAtom != atom.Span {
		return node, nil
	}
	_, exist := getAttrValExist(node, "data-custom")
	if !exist {
		return node, nil
	}
	_, err := d.Write([]byte(getAttrVal(node, "alt")))
	if err != nil {
		return nil, err
	}
	return nil, LoopChildren(node, d)
}

func assertStringWithRenderer(t *testing.T, input string, output string, n NodeRenderer) {
	assertPlaintextWithRenderer(t, input, ExactStringMatcher(output), n)
}
