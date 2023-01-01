package colly

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// ------------------------------------------------------------------------

// HTMLElement is the representation of a HTML tag.
type HTMLElement struct {
	Name       string             // Name is the name of the tag.
	Text       string             // Text is the inner text of the element.
	attributes []html.Attribute   // tag attributes
	Response   *Response          // Response is the Response object of the element's HTML document.
	DOM        *goquery.Selection // DOM is the goquery parsed DOM object of the page. DOM is relative to the current HTMLElement.
	Index      int                // Index stores the position of the current element within all the elements matched by an OnHTML callback.
}

// ------------------------------------------------------------------------

// NewHTMLElementFromSelectionNode creates a HTMLElement from a goquery.Selection Node.
func NewHTMLElementFromSelectionNode(resp *Response, s *goquery.Selection, n *html.Node, idx int) *HTMLElement {
	return &HTMLElement{
		Name:       n.Data,
		Response:   resp,
		Text:       goquery.NewDocumentFromNode(n).Text(),
		DOM:        s,
		Index:      idx,
		attributes: n.Attr,
	}
}

// ------------------------------------------------------------------------

// Attr returns the selected attribute of a HTMLElement or empty string if no attribute found.
func (h *HTMLElement) Attr(k string) string {
	for _, a := range h.attributes {
		if a.Key == k {
			return a.Val
		}
	}

	return ""
}

// ------------------------------------------------------------------------

// ChildText returns the concatenated and stripped text content of the matching elements.
func (h *HTMLElement) ChildText(goquerySelector string) string {
	return strings.TrimSpace(h.DOM.Find(goquerySelector).Text())
}

// ------------------------------------------------------------------------

// ChildTexts returns the stripped text content of all the matching elements.
func (h *HTMLElement) ChildTexts(goquerySelector string) []string {
	var texts = []string{}

	h.DOM.Find(goquerySelector).Each(func(_ int, s *goquery.Selection) {
		texts = append(texts, strings.TrimSpace(s.Text()))
	})

	return texts
}

// ------------------------------------------------------------------------

// ChildAttr returns the stripped text content of the first matching element's attribute.
func (h *HTMLElement) ChildAttr(goquerySelector, attrName string) string {
	if attr, ok := h.DOM.Find(goquerySelector).Attr(attrName); ok {
		return strings.TrimSpace(attr)
	}

	return ""
}

// ------------------------------------------------------------------------

// ChildAttrs returns the stripped text content of all the matching elements' attributes.
func (h *HTMLElement) ChildAttrs(goquerySelector, attrName string) []string {
	var attrs = []string{}

	h.DOM.Find(goquerySelector).Each(func(_ int, s *goquery.Selection) {
		if attr, ok := s.Attr(attrName); ok {
			attrs = append(attrs, strings.TrimSpace(attr))
		}
	})

	return attrs
}

// ------------------------------------------------------------------------

// ForEach iterates over the elements matched by the first argument
// and calls the callback function on every HTMLElement match.
func (h *HTMLElement) ForEach(goquerySelector string, callback func(int, *HTMLElement)) {
	var i int = 0

	h.DOM.Find(goquerySelector).Each(func(_ int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			callback(i, NewHTMLElementFromSelectionNode(h.Response, s, n, i))
			i++
		}
	})
}

// ------------------------------------------------------------------------

// ForEachWithBreak iterates over the elements matched by the first argument
// and calls the callback function on every HTMLElement match.
// It is identical to ForEach except that it is possible to break out of the loop
// by returning false in the callback function.
// It returns the current Selection object.
func (h *HTMLElement) ForEachWithBreak(goquerySelector string, callback func(int, *HTMLElement) bool) {
	var i int = 0

	h.DOM.Find(goquerySelector).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		for _, n := range s.Nodes {
			if callback(i, NewHTMLElementFromSelectionNode(h.Response, s, n, i)) {
				i++
				return true
			}
		}
		return false
	})
}
