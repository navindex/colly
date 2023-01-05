package colly

import (
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"golang.org/x/net/html"
)

// ------------------------------------------------------------------------

// XMLElement is the representation of a XML tag.
type XMLElement struct {
	Name       string    // Name is the name of the tag.
	Text       string    // Text is the inner text of the element.
	Response   *Response // Response is the Response object of the element's XML document.
	DOM        any       // DOM is the DOM object of the page. DOM is relative to the current XMLElement and is either a html.Node or xmlquery.Node.
	attributes any
	isHTML     bool
}

// ------------------------------------------------------------------------

// NewXMLElementFromHTMLNode creates a XMLElement from a html.Node.
func NewXMLElementFromHTMLNode(resp *Response, s *html.Node) *XMLElement {
	return &XMLElement{
		Name:       s.Data,
		Response:   resp,
		Text:       htmlquery.InnerText(s),
		DOM:        s,
		attributes: s.Attr,
		isHTML:     true,
	}
}

// ------------------------------------------------------------------------

// NewXMLElementFromXMLNode creates a XMLElement from a xmlquery.Node.
func NewXMLElementFromXMLNode(resp *Response, s *xmlquery.Node) *XMLElement {
	return &XMLElement{
		Name:       s.Data,
		Response:   resp,
		Text:       s.InnerText(),
		DOM:        s,
		attributes: s.Attr,
		isHTML:     false,
	}
}

// ------------------------------------------------------------------------

// Attr returns the selected attribute of a HTMLElement or empty string if no attribute found.
func (h *XMLElement) Attr(k string) string {
	if h.isHTML {
		for _, a := range h.attributes.([]html.Attribute) {
			if a.Key == k {
				return a.Val
			}
		}
	} else {
		for _, a := range h.attributes.([]xmlquery.Attr) {
			if a.Name.Local == k {
				return a.Value
			}
		}
	}

	return ""
}

// ------------------------------------------------------------------------

// ChildText returns the concatenated and stripped text content of the matching elements.
func (h *XMLElement) ChildText(xpathQuery string) string {
	if h.isHTML {
		child := htmlquery.FindOne(h.DOM.(*html.Node), xpathQuery)
		if child == nil {
			return ""
		}
		return strings.TrimSpace(htmlquery.InnerText(child))
	}
	child := xmlquery.FindOne(h.DOM.(*xmlquery.Node), xpathQuery)
	if child == nil {
		return ""
	}

	return strings.TrimSpace(child.InnerText())
}

// ------------------------------------------------------------------------

// ChildAttr returns the stripped text content of the first matching
// element's attribute.
func (h *XMLElement) ChildAttr(xpathQuery, attrName string) string {
	if h.isHTML {
		child := htmlquery.FindOne(h.DOM.(*html.Node), xpathQuery)
		if child != nil {
			for _, attr := range child.Attr {
				if attr.Key == attrName {
					return strings.TrimSpace(attr.Val)
				}
			}
		}
	} else {
		child := xmlquery.FindOne(h.DOM.(*xmlquery.Node), xpathQuery)
		if child != nil {
			for _, attr := range child.Attr {
				if attr.Name.Local == attrName {
					return strings.TrimSpace(attr.Value)
				}
			}
		}
	}

	return ""
}

// ------------------------------------------------------------------------

// ChildAttrs returns the stripped text content of all the matching elements' attributes.
func (h *XMLElement) ChildAttrs(xpathQuery, attrName string) []string {
	var attrs = []string{}

	if h.isHTML {
		for _, child := range htmlquery.Find(h.DOM.(*html.Node), xpathQuery) {
			for _, attr := range child.Attr {
				if attr.Key == attrName {
					attrs = append(attrs, strings.TrimSpace(attr.Val))
				}
			}
		}
	} else {
		xmlquery.FindEach(h.DOM.(*xmlquery.Node), xpathQuery, func(i int, child *xmlquery.Node) {
			for _, attr := range child.Attr {
				if attr.Name.Local == attrName {
					attrs = append(attrs, strings.TrimSpace(attr.Value))
				}
			}
		})
	}

	return attrs
}

// ------------------------------------------------------------------------

// ChildTexts returns an array of strings corresponding to child elements that match the xpath query.
// Each item in the array is the stripped text content of the corresponding matching child element.
func (h *XMLElement) ChildTexts(xpathQuery string) []string {
	var texts = []string{}

	if h.isHTML {
		for _, child := range htmlquery.Find(h.DOM.(*html.Node), xpathQuery) {
			texts = append(texts, strings.TrimSpace(htmlquery.InnerText(child)))
		}
	} else {
		xmlquery.FindEach(h.DOM.(*xmlquery.Node), xpathQuery, func(i int, child *xmlquery.Node) {
			texts = append(texts, strings.TrimSpace(child.InnerText()))
		})
	}

	return texts
}
