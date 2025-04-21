package engine

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/bluebrown/treasure-map/textfunc"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	minihtml "github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
	"sigs.k8s.io/yaml"

	"github.com/bluebrown/doktri/internal/fsys"
)

type Engine struct {
	src         string
	docs        string
	dist        string
	theme       string
	chromaStyle string
	markdown    goldmark.Markdown
	minifier    *minify.M
	meta        map[string]any
}

func New(options ...Option) Engine {
	// apply the options
	opts := Options{}
	for _, o := range options {
		o(&opts)
	}

	if opts.source == "" {
		opts.source = "."
	}

	if opts.dist == "" {
		opts.dist = filepath.Join(opts.source, "dist")
	}

	if opts.theme == "" {
		opts.theme = filepath.Join(opts.source, ".theme")
	}

	if opts.chromaStyle == "" {
		opts.chromaStyle = "dracula"
	}

	// md is the markdown rendering engine
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			emoji.Emoji,
			highlighting.NewHighlighting(
				highlighting.WithStyle(opts.chromaStyle),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
			),
			&frontmatter.Extender{
				Mode: frontmatter.SetMetadata,
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	m := minify.New()
	m.AddFunc("text/html", minihtml.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	return Engine{
		src:         opts.source,
		docs:        filepath.Join(opts.source, "docs"),
		dist:        opts.dist,
		theme:       opts.theme,
		chromaStyle: opts.chromaStyle,
		markdown:    md,
		minifier:    m,
		meta:        make(map[string]any),
	}
}

// create a funcmap to be used by the templates
func (e Engine) FuncMap() template.FuncMap {
	return NewFuncMapClosure(&e).FuncMap()
}

func (e Engine) MakeLayout(name string) (*template.Template, error) {
	var err error

	tpl := template.New("base.html")
	tpl.Funcs(textfunc.MapClosure(sprig.TxtFuncMap(), tpl)).Funcs(e.FuncMap())

	tpl, err = tpl.ParseFiles(
		filepath.Join(e.ThemeTemplatesDir(), "layouts", "base.html"),
		filepath.Join(e.ThemeTemplatesDir(), "layouts", name+".html"),
	)
	if err != nil {
		return nil, err
	}

	inc := filepath.Join(e.ThemeTemplatesDir(), "includes")

	// skip includes if there are none
	exists, err := fsys.PathExists(inc)
	if err != nil {
		return nil, fmt.Errorf("check includes exist: %w", err)
	}
	if !exists {
		return tpl, nil
	}
	empty, err := fsys.IsEmptyDir(inc)
	if err != nil {
		return nil, fmt.Errorf("check includes empty: %w", err)
	}
	if empty {
		return tpl, nil
	}

	return tpl.ParseGlob(filepath.Join(inc, "*"))
}

func (e Engine) SourceDir() string {
	return e.src
}

func (e Engine) DocsDir() string {
	return e.docs
}

func (e Engine) ThemeTemplatesDir() string {
	return filepath.Join(e.theme, "templates")
}

func (e Engine) DistDir() string {
	return e.dist
}

func (e Engine) ThemeAssetsDir() string {
	return filepath.Join(e.theme, "assets")
}

func (e Engine) ExtraAssetsDir() string {
	return filepath.Join(e.src, "assets")
}

func (e Engine) MetaPath() string {
	return filepath.Join(e.src, "meta.yaml")
}

func (e Engine) Meta() map[string]any {
	return e.meta
}

func (e Engine) Run() error {
	var err error
	// reset the dist dir
	if err := os.RemoveAll(e.DistDir()); err != nil {
		return fmt.Errorf("clean dist: %w", err)
	}

	if err := os.MkdirAll(e.DistDir(), 0755); err != nil {
		return fmt.Errorf("create dist: %w", err)
	}

	// initialize the walker
	walker := TreeWalker{mini: e.minifier}

	// read the meta file
	exists, err := fsys.PathExists(e.MetaPath())
	if err != nil {
		return err
	}
	if exists {
		b, err := os.ReadFile(e.MetaPath())
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(b, &e.meta); err != nil {
			return err
		}
	}

	walker.dirTpl, err = e.MakeLayout("dir")
	if err != nil {
		return fmt.Errorf("read dir tpl: %w", err)
	}

	walker.fileTpl, err = e.MakeLayout("file")
	if err != nil {
		return fmt.Errorf("read file tpl: %w", err)
	}

	walker.srcFS = os.DirFS(e.DocsDir())
	walker.distPath = e.DistDir()

	// build a new tree from the src FS
	treeRoot, err := buildTree(walker.srcFS)
	if err != nil {
		return fmt.Errorf("walk: %w", err)
	}

	// sort the tree by date and then walk it
	err = walker.RenderWalk(treeRoot.SortDate(SortDirectionDescending))
	if err != nil {
		return fmt.Errorf("walk: %w", err)
	}

	// dist assets is the location in dist to put additional assets
	distAssets := filepath.Join(e.DistDir(), filepath.Base(e.ThemeAssetsDir()))

	// copy first the theme assets and then the extra assets
	for _, p := range []string{e.ThemeAssetsDir(), e.ExtraAssetsDir()} {
		// copy assets to dist assets if any
		ok, err := fsys.PathExists(p)
		if err != nil {
			return fmt.Errorf("check assets: %w", err)
		}
		if ok {
			if err := e.copyAssets(p, distAssets); err != nil {
				return fmt.Errorf("copy assets: %w", err)
			}
		}
	}

	// ensure dist assets dir, in case it has not been created by
	// the themes or extra asset copy
	if err := os.MkdirAll(distAssets, 0755); err != nil {
		return fmt.Errorf("create assets dir: %w", err)
	}

	f, err := os.Create(filepath.Join(e.DistDir(), "assets", "chroma.css"))
	if err != nil {
		return fmt.Errorf("create chroma.css: %w", err)
	}

	// generate the styles
	w := e.minifier.Writer("text/css", f)
	errs := make(Errors, 0, 3)

	if err = GenerateStyles(w, e.chromaStyle); err != nil {
		errs = append(errs, fmt.Errorf("generate styles: %w", err))
	}

	if err := w.Close(); err != nil {
		errs = append(errs, fmt.Errorf("minify chroma.css: %w", err))
	}

	if err := f.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close chroma.css: %w", err))
	}

	if len(errs) != 0 {
		return errs
	}

	return nil

}

type TreeWalker struct {
	distPath string
	srcFS    fs.FS
	dirTpl   *template.Template
	fileTpl  *template.Template
	mini     *minify.M
}

func (tw TreeWalker) RenderWalk(node *TreeNode) error {
	if node.IsRoot {
		fmt.Printf("\n%-25s   %s\n", "PARENT", "PAGE")
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	} else {
		fmt.Printf("%-25s < %s\n", node.Parent.Path(), node.Name())
	}

	// render template
	var (
		t *template.Template
		p string
	)

	// NOTE: do not use node.Path(), to get the path since we should use the os
	// specific path separator and node.Path() returns forward slashes as its
	// meant to be used as web link
	if node.IsLeaf {
		t = tw.fileTpl
		p = filepath.Join(tw.distPath, filepath.Dir(node.SourcePath), NormalizeMdName(filepath.Base(node.SourcePath)), "index.html")
	} else {
		t = tw.dirTpl
		p = filepath.Join(tw.distPath, node.SourcePath, "index.html")
	}

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("create dist dir: %w", err)
	}

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("create distr file: %w", err)
	}

	// render the template to a buffer
	buf := new(bytes.Buffer)
	err = t.Execute(buf, node)
	if err != nil {
		f.Close()
		return fmt.Errorf("exec template: %w", err)
	}

	// minify it
	if err := tw.mini.Minify("text/html", f, buf); err != nil {
		f.Close()
		return fmt.Errorf("minify html: %w", err)
	}

	// don't use defer for the file.Close, otherwise we have a lot of open file
	// descriptors until the whole tree is handled this is because the function
	// doesn't return until the children are handled.
	if err := f.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	// repeat for children recursively
	for _, c := range node.Children {
		if err := tw.RenderWalk(c); err != nil {
			return err
		}
	}
	return nil
}

func (e Engine) copyAssets(root, target string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("assets walk: %w", err)
		}

		outPath := filepath.Join(target, strings.TrimPrefix(path, root))

		if d.IsDir() {
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return fmt.Errorf("assets copy: create dir: %w", err)
			}
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("assets copy: open src: %w", err)
		}
		defer src.Close()

		dst, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("assets copy: open dst: %w", err)
		}

		defer dst.Close()

		if _, params, m := e.minifier.Match(mime.TypeByExtension(filepath.Ext(path))); m != nil {
			if err := m.Minify(e.minifier, dst, src, params); err != nil {
				return fmt.Errorf("assets minify: %s -> %s: %w", path, outPath, err)
			}
		} else {
			if _, err := io.Copy(dst, src); err != nil {
				return fmt.Errorf("assets copy: %s -> %s: %w", path, outPath, err)
			}
		}

		return nil
	})
}

// generates CSS styles with the given theme or fallback and write them to the writer
func GenerateStyles(dist io.Writer, theme string, opts ...chromahtml.Option) error {
	style := styles.Get(theme)
	if style == nil {
		style = styles.Fallback
	}
	return chromahtml.New(opts...).WriteCSS(dist, style)
}

type Errors []error

func (ee Errors) Error() string {
	var buf bytes.Buffer
	for i, e := range ee {
		if i != 0 {
			buf.WriteString(": ")
		}
		buf.WriteString(e.Error())
	}
	return buf.String()
}
