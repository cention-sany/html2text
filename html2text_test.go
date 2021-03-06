package html2text

import (
	"regexp"
	"testing"
)

func TestStrippingWhitespace(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"test text",
			"test text",
		},
		{
			"  \ttext\ntext\n",
			"text text",
		},
		{
			"  \na \n\t \n \n a \t",
			"a a",
		},
		{
			"test        text",
			"test text",
		},
		{
			"test&nbsp;&nbsp;&nbsp; text&nbsp;",
			"test    text",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}
}

func TestParagraphsAndBreaks(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"Test text",
			"Test text",
		},
		{
			"Test text<br>",
			"Test text",
		},
		{
			"Test text<br>Test",
			"Test text\nTest",
		},
		{
			"<p>Test text</p>",
			"Test text",
		},
		{
			"<p>Test text</p><p>Test text</p>",
			"Test text\n\nTest text",
		},
		{
			"\n<p>Test text</p>\n\n\n\t<p>Test text</p>\n",
			"Test text\n\nTest text",
		},
		{
			"\n<p>Test text<br/>Test text</p>\n",
			"Test text\nTest text",
		},
		{
			"\n<p>Test text<br> \tTest text<br></p>\n",
			"Test text\nTest text",
		},
		{
			"Test text<br><BR />Test text",
			"Test text\n\nTest text",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}
}

func TestTables(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<table><tr><td></td><td></td></tr></table>",
			"",
		},
		{
			"<table><tr><td>cell1</td><td>cell2</td></tr></table>",
			"cell1 cell2",
		},
		{
			"<table><tr><td>row1</td></tr><tr><td>row2</td></tr></table>",
			"row1\nrow2",
		},
		{
			`<table>
			   <tr><td>cell1-1</td><td>cell1-2</td></tr>
			   <tr><td>cell2-1</td><td>cell2-2</td></tr>
			</table>`,
			"cell1-1 cell1-2\ncell2-1 cell2-2",
		},
		{
			"_<table><tr><td>cell</td></tr></table>_",
			"_\n\ncell\n\n_",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}
}

func TestStrippingLists(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<ul></ul>",
			"",
		},
		{
			"<ul><li>item</li></ul>_",
			"* item\n\n_",
		},
		{
			"<li class='123'>item 1</li> <li>item 2</li>\n_",
			"* item 1\n* item 2\n_",
		},
		{
			"<li>item 1</li> \t\n <li>item 2</li> <li> item 3</li>\n_",
			"* item 1\n* item 2\n* item 3\n_",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}
}

func TestLinks(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			`<a></a>`,
			``,
		},
		{
			`<a href=""></a>`,
			``,
		},
		{
			`<a href="http://example.com/"></a>`,
			`( http://example.com/ )`,
		},
		{
			`<a href="">Link</a>`,
			`Link`,
		},
		{
			`<a href="http://example.com/">Link</a>`,
			`Link ( http://example.com/ )`,
		},
		{
			`<a href="http://example.com/"><span class="a">Link</span></a>`,
			`Link ( http://example.com/ )`,
		},
		{
			"<a href='http://example.com/'>\n\t<span class='a'>Link</span>\n\t</a>",
			`Link ( http://example.com/ )`,
		},
		{
			"<a href='mailto:contact@example.org'>Contact Us</a>",
			`Contact Us ( contact@example.org )`,
		},
		{
			"<a href=\"http://example.com:80/~user?aaa=bb&amp;c=d,e,f#foo\">Link</a>",
			`Link ( http://example.com:80/~user?aaa=bb&c=d,e,f#foo )`,
		},
		{
			"<a title='title' href=\"http://example.com/\">Link</a>",
			`Link ( http://example.com/ )`,
		},
		{
			"<a href=\"   http://example.com/ \"> Link </a>",
			`Link ( http://example.com/ )`,
		},
		{
			"<a href=\"http://example.com/a/\">Link A</a> <a href=\"http://example.com/b/\">Link B</a>",
			`Link A ( http://example.com/a/ ) Link B ( http://example.com/b/ )`,
		},
		{
			"<a href=\"%%LINK%%\">Link</a>",
			`Link ( %%LINK%% )`,
		},
		{
			"<a href=\"[LINK]\">Link</a>",
			`Link ( [LINK] )`,
		},
		{
			"<a href=\"{LINK}\">Link</a>",
			`Link ( {LINK} )`,
		},
		{
			"<a href=\"[[!unsubscribe]]\">Link</a>",
			`Link ( [[!unsubscribe]] )`,
		},
		{
			"<p>This is <a href=\"http://www.google.com\" >link1</a> and <a href=\"http://www.google.com\" >link2 </a> is next.</p>",
			`This is link1 ( http://www.google.com ) and link2 ( http://www.google.com ) is next.`,
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}
}

func TestImageAltTags(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			`<img />`,
			``,
		},
		{
			`<img src="http://example.ru/hello.jpg" />`,
			``,
		},
		{
			`<img alt="Example"/>`,
			``,
		},
		{
			`<img src="http://example.ru/hello.jpg" alt="Example"/>`,
			``,
		},
		// Images do matter if they are in a link
		{
			`<a href="http://example.com/"><img src="http://example.ru/hello.jpg" alt="Example"/></a>`,
			`Example ( http://example.com/ )`,
		},
		{
			`<a href="http://example.com/"><img src="http://example.ru/hello.jpg" alt="Example"></a>`,
			`Example ( http://example.com/ )`,
		},
		{
			`<a href='http://example.com/'><img src='http://example.ru/hello.jpg' alt='Example'/></a>`,
			`Example ( http://example.com/ )`,
		},
		{
			`<a href='http://example.com/'><img src='http://example.ru/hello.jpg' alt='Example'></a>`,
			`Example ( http://example.com/ )`,
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}
}

func TestHeadings(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<h1>Test</h1>",
			"****\nTest\n****",
		},
		{
			"\t<h1>\nTest</h1> ",
			"****\nTest\n****",
		},
		{
			"\t<h1>\nTest line 1<br>Test 2</h1> ",
			"***********\nTest line 1\nTest 2\n***********",
		},
		{
			"<h1>Test</h1> <h1>Test</h1>",
			"****\nTest\n****\n\n****\nTest\n****",
		},
		{
			"<h2>Test</h2>",
			"----\nTest\n----",
		},
		{
			"<h1><a href='http://example.com/'>Test</a></h1>",
			"****************************\nTest ( http://example.com/ )\n****************************",
		},
		{
			"<h3> <span class='a'>Test </span></h3>",
			"Test\n----",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}

}

func TestBold(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<b>Test</b>",
			"*Test*",
		},
		{
			"\t<b>Test</b> ",
			"*Test*",
		},
		{
			"\t<b>Test line 1<br>Test 2</b> ",
			"*Test line 1\nTest 2*",
		},
		{
			"<b>Test</b> <b>Test</b>",
			"*Test* *Test*",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}

}

func TestDiv(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<div>Test</div>",
			"Test",
		},
		{
			"\t<div>Test</div> ",
			"Test",
		},
		{
			"<div>Test line 1<div>Test 2</div></div>",
			"Test line 1\nTest 2",
		},
		{
			"Test 1<div>Test 2</div> <div>Test 3</div>Test 4",
			"Test 1\nTest 2\nTest 3\nTest 4",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}

}

func TestBlockquotes(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<div>level 0<blockquote>level 1<br><blockquote>level 2</blockquote>level 1</blockquote><div>level 0</div></div>",
			"level 0\n> \n> level 1\n> \n>> level 2\n> \n> level 1\n\nlevel 0",
		},
		{
			"<blockquote>Test</blockquote>Test",
			"> \n> Test\n\nTest",
		},
		{
			"\t<blockquote> \nTest<br></blockquote> ",
			"> \n> Test\n>",
		},
		{
			"\t<blockquote> \nTest line 1<br>Test 2</blockquote> ",
			"> \n> Test line 1\n> Test 2",
		},
		{
			"<blockquote>Test</blockquote> <blockquote>Test</blockquote> Other Test",
			"> \n> Test\n\n> \n> Test\n\nOther Test",
		},
		{
			"<blockquote>Lorem ipsum Commodo id consectetur pariatur ea occaecat minim aliqua ad sit consequat quis ex commodo Duis incididunt eu mollit consectetur fugiat voluptate dolore in pariatur in commodo occaecat Ut occaecat velit esse labore aute quis commodo non sit dolore officia Excepteur cillum amet cupidatat culpa velit labore ullamco dolore mollit elit in aliqua dolor irure do</blockquote>",
			"> \n> Lorem ipsum Commodo id consectetur pariatur ea occaecat minim aliqua ad\n> sit consequat quis ex commodo Duis incididunt eu mollit consectetur fugiat\n> voluptate dolore in pariatur in commodo occaecat Ut occaecat velit esse\n> labore aute quis commodo non sit dolore officia Excepteur cillum amet\n> cupidatat culpa velit labore ullamco dolore mollit elit in aliqua dolor\n> irure do",
		},
		{
			"<blockquote>Lorem<b>ipsum</b><b>Commodo</b><b>id</b><b>consectetur</b><b>pariatur</b><b>ea</b><b>occaecat</b><b>minim</b><b>aliqua</b><b>ad</b><b>sit</b><b>consequat</b><b>quis</b><b>ex</b><b>commodo</b><b>Duis</b><b>incididunt</b><b>eu</b><b>mollit</b><b>consectetur</b><b>fugiat</b><b>voluptate</b><b>dolore</b><b>in</b><b>pariatur</b><b>in</b><b>commodo</b><b>occaecat</b><b>Ut</b><b>occaecat</b><b>velit</b><b>esse</b><b>labore</b><b>aute</b><b>quis</b><b>commodo</b><b>non</b><b>sit</b><b>dolore</b><b>officia</b><b>Excepteur</b><b>cillum</b><b>amet</b><b>cupidatat</b><b>culpa</b><b>velit</b><b>labore</b><b>ullamco</b><b>dolore</b><b>mollit</b><b>elit</b><b>in</b><b>aliqua</b><b>dolor</b><b>irure</b><b>do</b></blockquote>",
			"> \n> Lorem *ipsum* *Commodo* *id* *consectetur* *pariatur* *ea* *occaecat* *minim*\n> *aliqua* *ad* *sit* *consequat* *quis* *ex* *commodo* *Duis* *incididunt* *eu*\n> *mollit* *consectetur* *fugiat* *voluptate* *dolore* *in* *pariatur* *in* *commodo*\n> *occaecat* *Ut* *occaecat* *velit* *esse* *labore* *aute* *quis* *commodo*\n> *non* *sit* *dolore* *officia* *Excepteur* *cillum* *amet* *cupidatat* *culpa*\n> *velit* *labore* *ullamco* *dolore* *mollit* *elit* *in* *aliqua* *dolor* *irure*\n> *do*",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}

}

func TestIgnoreStylesScriptsHead(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<style>Test</style>",
			"",
		},
		{
			"<style type=\"text/css\">body { color: #fff; }</style>",
			"",
		},
		{
			"<link rel=\"stylesheet\" href=\"main.css\">",
			"",
		},
		{
			"<script>Test</script>",
			"",
		},
		{
			"<script src=\"main.js\"></script>",
			"",
		},
		{
			"<script type=\"text/javascript\" src=\"main.js\"></script>",
			"",
		},
		{
			"<script type=\"text/javascript\">Test</script>",
			"",
		},
		{
			"<script type=\"text/ng-template\" id=\"template.html\"><a href=\"http://google.com\">Google</a></script>",
			"",
		},
		{
			"<script type=\"bla-bla-bla\" id=\"template.html\">Test</script>",
			"",
		},
		{
			`<html><head><title>Title</title></head><body></body></html>`,
			"",
		},
	}

	for _, testCase := range testCases {
		assertString(t, testCase.input, testCase.output)
	}
}

func TestText(t *testing.T) {
	testCases := []struct {
		input string
		expr  string
	}{
		{
			`<li>
		  <a href="/new" data-ga-click="Header, create new repository, icon:repo"><span class="octicon octicon-repo"></span> New repository</a>
		</li>`,
			`\* New repository \( /new \)`,
		},
		{
			`hi

			<br>
	
	hello <a href="https://google.com">google</a>
	<br><br>
	test<p>List:</p>

	<ul>
		<li><a href="foo">Foo</a></li>
		<li><a href="http://www.microshwhat.com/bar/soapy">Barsoap</a></li>
        <li>Baz</li>
	</ul>
`,
			`hi
hello google \( https://google.com \)

test

List:

\* Foo \( foo \)
\* Barsoap \( http://www.microshwhat.com/bar/soapy \)
\* Baz`,
		},
		// Malformed input html.
		{
			`hi

			hello <a href="https://google.com">google</a>

			test<p>List:</p>

			<ul>
				<li><a href="foo">Foo</a>
				<li><a href="/
		                bar/baz">Bar</a>
		        <li>Baz</li>
			</ul>
		`,
			`hi hello google \( https://google.com \) test

List:

\* Foo \( foo \)
\* Bar \( /\n[ \t]+bar/baz \)
\* Baz`,
		},
	}

	for _, testCase := range testCases {
		assertRegexp(t, testCase.input, testCase.expr)
	}
}

type StringMatcher interface {
	MatchString(string) bool
	String() string
}

type RegexpStringMatcher string

func (m RegexpStringMatcher) MatchString(str string) bool {
	return regexp.MustCompile(string(m)).MatchString(str)
}
func (m RegexpStringMatcher) String() string {
	return string(m)
}

type ExactStringMatcher string

func (m ExactStringMatcher) MatchString(str string) bool {
	return string(m) == str
}
func (m ExactStringMatcher) String() string {
	return string(m)
}

func assertRegexp(t *testing.T, input string, outputRE string) {
	assertPlaintext(t, input, RegexpStringMatcher(outputRE))
}

func assertString(t *testing.T, input string, output string) {
	assertPlaintext(t, input, ExactStringMatcher(output))
}

func assertPlaintext(t *testing.T, input string, matcher StringMatcher) {
	assertPlaintextWithRenderer(t, input, matcher, nil)
}

func assertPlaintextWithRenderer(t *testing.T, input string, matcher StringMatcher, n NodeRenderer) {
	text, err := FromStringWithRenderer(input, n)
	if err != nil {
		t.Error(err)
	}
	if !matcher.MatchString(text) {
		t.Errorf("Input did not match expression\n"+
			"Input:\n>>>>\n%s\n<<<<\n\n"+
			"Output:\n>>>>\n%s\n<<<<\n\n"+
			"Expected output:\n>>>>\n%s\n<<<<\n\n",
			input, text, matcher.String())
	} else {
		t.Logf("input:\n\n%s\n\n\n\noutput:\n\n%s\n", input, text)
	}
}
