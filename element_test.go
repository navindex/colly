package colly

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// ------------------------------------------------------------------------

func setupXMLElementTestCase() (*Response, *html.Node) {
	// Borrowed from http://infohost.nmt.edu/tcc/help/pubs/xhtml/example.html
	// Added attributes to the `<li>` tags for testing purposes
	htmlPage := `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN"
 "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
  <head>
    <title>Your page title here</title>
  </head>
  <body>
    <h1>Your major heading here</h1>
    <p>
      This is a regular text paragraph.
    </p>
    <ul>
      <li class="list-item-1">
        First bullet of a bullet list.
      </li>
      <li class="list-item-2">
        This is the <em>second</em> bullet.
      </li>
    </ul>
  </body>
</html>
`
	resp := &Response{
		Resp: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(htmlPage)),
		},
	}
	doc, _ := htmlquery.Parse(strings.NewReader(htmlPage))

	return resp, doc
}

// ------------------------------------------------------------------------

func TestXMLElement_Attr(t *testing.T) {
	resp, doc := setupXMLElementTestCase()
	xmlNode := htmlquery.FindOne(doc, "/html")
	xmlElem := NewXMLElementFromHTMLNode(resp, xmlNode)

	if xmlElem.Attr("xmlns") != "http://www.w3.org/1999/xhtml" {
		t.Fatalf("failed xmlns attribute test: %v != http://www.w3.org/1999/xhtml", xmlElem.Attr("xmlns"))
	}

	if xmlElem.Attr("xml:lang") != "en" {
		t.Fatalf("failed lang attribute test: %v != en", xmlElem.Attr("lang"))
	}
}

// ------------------------------------------------------------------------

func TestXMLElement_ChildText(t *testing.T) {
	resp, doc := setupXMLElementTestCase()
	xmlNode := htmlquery.FindOne(doc, "/html")
	xmlElem := NewXMLElementFromHTMLNode(resp, xmlNode)

	if text := xmlElem.ChildText("//p"); text != "This is a regular text paragraph." {
		t.Fatalf("failed child tag test: %v != This is a regular text paragraph.", text)
	}
	if text := xmlElem.ChildText("//dl"); text != "" {
		t.Fatalf("failed child tag test: %v != \"\"", text)
	}
}

// ------------------------------------------------------------------------

func TestXMLElement_ChildTexts(t *testing.T) {
	resp, doc := setupXMLElementTestCase()
	xmlNode := htmlquery.FindOne(doc, "/html")
	xmlElem := NewXMLElementFromHTMLNode(resp, xmlNode)
	expected := []string{"First bullet of a bullet list.", "This is the second bullet."}

	if texts := xmlElem.ChildTexts("//li"); reflect.DeepEqual(texts, expected) == false {
		t.Fatalf("failed child tags test: %v != %v", texts, expected)
	}

	if texts := xmlElem.ChildTexts("//dl"); reflect.DeepEqual(texts, make([]string, 0)) == false {
		t.Fatalf("failed child tag test: %v != \"\"", texts)
	}
}

// ------------------------------------------------------------------------

func TestXMLElement_ChildAttr(t *testing.T) {
	resp, doc := setupXMLElementTestCase()
	xmlNode := htmlquery.FindOne(doc, "/html")
	xmlElem := NewXMLElementFromHTMLNode(resp, xmlNode)

	if attr := xmlElem.ChildAttr("/body/ul/li[1]", "class"); attr != "list-item-1" {
		t.Fatalf("failed child attribute test: %v != list-item-1", attr)
	}
	if attr := xmlElem.ChildAttr("/body/ul/li[2]", "class"); attr != "list-item-2" {
		t.Fatalf("failed child attribute test: %v != list-item-2", attr)
	}
}

// ------------------------------------------------------------------------

func TestXMLElement_ChildAttrs(t *testing.T) {
	resp, doc := setupXMLElementTestCase()
	xmlNode := htmlquery.FindOne(doc, "/html")
	xmlElem := NewXMLElementFromHTMLNode(resp, xmlNode)

	attrs := xmlElem.ChildAttrs("/body/ul/li", "class")
	if len(attrs) != 2 {
		t.Fatalf("failed child attributes length test: %d != 2", len(attrs))
	}

	for _, attr := range attrs {
		if !(attr == "list-item-1" || attr == "list-item-2") {
			t.Fatalf("failed child attributes values test: %s != list-item-(1 or 2)", attr)
		}
	}
}
