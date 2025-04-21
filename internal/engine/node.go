package engine

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	// the context path can be used if the page is not hosted at the domain root
	CONTEXT_PATH = "/"
	// the post author is used globally for all posts
	POST_AUTHOR = "Anonymous"
)

// normalize the string by stripping the date prefix and .md suffix
func NormalizeMdName(s string) string {
	return strings.TrimSuffix(s, ".md")[11:]
}

type SortDirection string

const (
	SortDirectionAscending  SortDirection = "asc"
	SortDirectionDescending SortDirection = "desc"
)

// the cache holds values that dont change
// they will be set on first call of the corresponding method
type nodeCache struct {
	date  *time.Time
	path  string
	name  string
	title string
}

type TreeNode struct {
	// the source fs
	fs fs.FS
	// the path withing the source fs
	SourcePath string
	// the raw entry
	Entry fs.DirEntry
	// true if its the top level root node
	IsRoot bool
	// true if node does not have children aka is a file and not a dir
	IsLeaf bool
	// pointer to the parent node. Is nil for the root node
	Parent *TreeNode
	// slice of children. Always empty for leave nodes
	Children TreeNodeList
	// pointer to root node. will point to itself for the treeRoot
	// this makes it more easy to use it in templates
	Root *TreeNode
	// the cache is an internal struct to hold values
	// that are cached when methods are called
	cache *nodeCache
}

// get the content for this node by reading the source file this is done in this
// method so we don't need to read all files into memory, at once. Each piece of
// content can be read lazily when its actually needed because a given template
// wants to use it. if its not a leaf note, the content of the index.md in the
// given dir is returned if possible. Otherwise the byte slice will have len 0
func (n *TreeNode) Content() []byte {
	if !n.IsLeaf {
		b, err := fs.ReadFile(n.fs, filepath.Join(n.SourcePath, "index.md"))
		if err != nil {
			return []byte{}
		}
		return b
	}
	b, err := fs.ReadFile(n.fs, n.SourcePath)
	if err != nil {
		panic("failed to read content")
	}
	return b
}

// return the normalized path as its used on the web page.
// Meaning it does not point to the nodes source and it will
// always point to a directory because even leaf nodes are created
// as index.html under a directory with the leaf nodes name
func (n *TreeNode) Path() string {
	if n.cache.path == "" {
		if n.IsRoot {
			n.cache.path = CONTEXT_PATH
		} else {
			s := n.SourcePath
			if n.IsLeaf {
				s = filepath.Join(filepath.Dir(s), NormalizeMdName(filepath.Base(s)))
			}
			n.cache.path = CONTEXT_PATH + filepath.ToSlash(s) + "/"
		}
	}
	return n.cache.path
}

// Return the normalized name as its used on a web page.
// this will strip the date prefix and .md suffix
func (n *TreeNode) Name() string {
	if n.cache.name != "" {
		return n.cache.name
	}
	if n.IsRoot {
		n.cache.name = "home"
	} else {
		n.cache.name = filepath.Base(n.SourcePath)
		if n.IsLeaf {
			// remove the date prefix and .md suffix, if its a leaf
			n.cache.name = NormalizeMdName(n.cache.name)
		}
	}
	return n.cache.name
}

// return the normalized name as human readable title.
// dashes are replaced with spaces and the word tokens are title cased
func (n *TreeNode) Title() string {
	if n.cache.title == "" {
		// TODO: handle abbreviations
		// name := n.Name()
		// if its 3 or less, its probably an abbreviation
		// if len(name) <= 3 {
		// 	return strings.ToUpper(name)
		// }
		caser := cases.Title(language.English)
		n.cache.title = caser.String(strings.ReplaceAll(n.Name(), "-", " "))
	}
	return n.cache.title
}

// convenience function to get a nodes siblings this is the same as getting the
// parents children, filtering itself out. will panic when called on the root
// node, since a root has no parent and therefore no siblings
func (n *TreeNode) Siblings() TreeNodeList {
	if n.IsRoot {
		panic("root node cannot have siblings")
	}
	ss := make([]*TreeNode, 0, len(n.Parent.Children)-1)
	for _, c := range n.Parent.Children {
		if c == n {
			continue
		}
		ss = append(ss, c)
	}
	return ss
}

// get the first child of this node. Returns nil of node has no children
func (n *TreeNode) FirstChild() *TreeNode {
	if len(n.Children) < 1 {
		return nil
	}
	return n.Children[0]
}

// get the next sibling. Panics if called on the root node
// returns nil of node has no next sibling
func (n *TreeNode) NextSibling() *TreeNode {
	if n.IsRoot {
		panic("root node cannot have siblings")
	}
	var ni int
	for i, c := range n.Parent.Children {
		if c == n {
			ni = i + 1
			break
		}
	}
	if len(n.Parent.Children) < ni+1 {
		return nil
	}
	return n.Parent.Children[ni]
}

// get the previous sibling. Panics if called on the root node
// returns nil of node has no previous sibling
func (n *TreeNode) PreviousSibling() *TreeNode {
	if n.IsRoot {
		panic("root node cannot have siblings")
	}
	var ni int
	for i, c := range n.Parent.Children {
		if c == n {
			ni = i - 1
			break
		}
	}
	if ni < 0 {
		return nil
	}
	return n.Parent.Children[ni]
}

// traverse the tree starting from this node to all leafs, and sort
// the children of each node
func (n *TreeNode) SortDate(direction SortDirection) *TreeNode {
	// first sort the children
	n.Children.SortDate(direction)
	// and repeat recursively
	for _, c := range n.Children {
		c.SortDate(direction)
	}
	return n
}

// return the creation date of the node. For leafs the date will be inferred
// from the date-suffix of the source file i.e. 2022-03-05-myfile.md.
// for non-leafs (folders) the date of the oldest children will be used.
// if the node is non-leaf and has no children, using oldest is not possible
// in that case it will fallback to using the sourceFiles modtime.
func (n *TreeNode) Date() time.Time {
	if n.cache.date != nil {
		return *n.cache.date
	}

	var t time.Time

	if !n.IsLeaf {
		if oldest := n.Children.Oldest(); oldest != nil {
			t = oldest.Date()
		} else {
			info, err := n.Entry.Info()
			if err != nil {
				panic(fmt.Sprintf("fileinfo: %s", err.Error()))
			}
			t = info.ModTime()
		}
	} else {
		var err error
		ds := filepath.Base(n.SourcePath)[:10]
		t, err = time.Parse("2006-01-02", ds)
		if err != nil {
			panic(fmt.Sprintf("failed to parse date: %s", ds))
		}
	}

	// cache the value for the next lookup
	n.cache.date = &t

	return t
}

// return the author of the node
// this is currently set to a static value
// since we do not use front matter
func (n *TreeNode) Author() string {
	return POST_AUTHOR
}

// return true if node has children
func (n *TreeNode) HasChildren() bool {
	return len(n.Children) > 0
}

// return true if node has siblings
func (n *TreeNode) HasSiblings() bool {
	return !n.IsRoot && (len(n.Siblings()) > 0)
}

type TreeNodeList []*TreeNode

// sorts the child list in place. Making the sort permanent
func (tc TreeNodeList) SortDate(direction SortDirection) TreeNodeList {
	sort.SliceStable(tc, func(i, j int) bool {
		var a, b int
		if direction == SortDirectionAscending {
			a = i
			b = j
		} else {
			a = j
			b = i
		}
		return tc[a].Date().Before(tc[b].Date())
	})
	// returns itself for chaining
	return tc
}

// get the oldest (Date) child. Since this is calling
// Date on each children, and the Date function calls oldest,
// for non-leaf nodes, this will recurse the tree until leafs
// hare found and propagate their date upwards to the parent
func (tc TreeNodeList) Oldest() *TreeNode {
	if len(tc) == 0 {
		return nil
	}
	shadow := make(TreeNodeList, len(tc))
	copy(shadow, tc)
	sort.SliceStable(shadow, func(i, j int) bool {
		return shadow[i].Date().Before(shadow[j].Date())
	})
	return shadow[0]
}

// get all leafs
func (tc TreeNodeList) Leafs() TreeNodeList {
	l := make(TreeNodeList, 0, len(tc))
	for _, c := range tc {
		if !c.IsLeaf {
			continue
		}
		l = append(l, c)
	}
	return l
}
