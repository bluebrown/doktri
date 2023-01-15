package engine

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/toc"
)

// closes over the engine to provide specialized functions that can be used
// inside the templates. The methods of this struct are available as lowercase
// functions inside the funcmap
type FuncMapClosure struct {
	e *Engine
}

// create a new closure funcmap instance
func NewFuncMapClosure(e *Engine) *FuncMapClosure {
	return &FuncMapClosure{e}
}

// generate the funcmap to be used by the templates
func (fmc *FuncMapClosure) FuncMap() template.FuncMap {
	return map[string]any{
		"meta":    fmc.Meta(),
		"render":  fmc.Render(),
		"toc":     fmc.Toc(),
		"excerpt": fmc.Excerpt(),
		"link":    fmc.Link(),
	}
}

// retrieve the structured content of the meta.yaml
func (fmc *FuncMapClosure) Meta() func() map[string]any {
	return func() map[string]any {
		return fmc.e.Meta()
	}
}

// convert the given raw markdown bytes to html
func (fmc *FuncMapClosure) Render() func(b []byte) string {
	return func(b []byte) string {
		var buf bytes.Buffer
		if err := fmc.e.markdown.Convert(b, &buf); err != nil {
			panic(err)
		}
		return buf.String()
	}
}

// generate a table of contents from the raw markdown bytes. the toc is returned
// as html ul element
func (fmc *FuncMapClosure) Toc() func(b []byte) string {
	return func(b []byte) string {
		doc := fmc.e.markdown.Parser().Parse(text.NewReader(b))
		tree, err := toc.Inspect(doc, b)
		if err != nil {
			panic(err)
		}
		list := toc.RenderList(tree)

		// the first child is the first list item
		n := list.FirstChild()
		if n == nil {
			return ""
		}

		// the first child of that is the anchor
		n = n.FirstChild()
		if n == nil {
			return ""
		}

		// the sibling of that is the nested ul
		n = n.NextSibling()
		if n == nil {
			return ""
		}

		buf := new(bytes.Buffer)
		if err := fmc.e.markdown.Renderer().Render(buf, b, n); err != nil {
			panic(err)
		}

		return buf.String()
	}
}

// generate an expert in form of an html paragraph. the paragraph will be the
// first paragraph of the raw markdown. if the markdown has no paragraphs, the
// returned string is empty
func (fmc *FuncMapClosure) Excerpt() func(b []byte) string {
	return func(b []byte) string {
		// TODO: find out if paragraph parser can be used
		doc := fmc.e.markdown.Parser().Parse(text.NewReader(b))
		firstParagraph := doc.FirstChild().NextSibling()
		if firstParagraph == nil {
			return ""
		}
		buf := new(bytes.Buffer)
		if err := fmc.e.markdown.Renderer().Render(buf, b, firstParagraph); err != nil {
			panic(err)
		}
		return buf.String()
	}
}

// create an html link tag resolving to the assets dir. This is useful because
// it takes the context path into consideration. When the context path is
// changed the generated links will change accordingly. Use this to link local
// assets in your templates
func (fmc *FuncMapClosure) Link() func(href, rel string) string {
	return func(href, rel string) string {
		return fmt.Sprintf(`<link rel="%s" href="%s%s/%s">`, rel, CONTEXT_PATH, "assets", href)
	}
}
