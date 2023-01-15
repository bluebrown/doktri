package engine

import (
	"fmt"
	"io/fs"
)

func buildTree(srcFS fs.FS) (*TreeNode, error) {
	var (
		treeRoot      *TreeNode
		currentBranch *TreeNode
	)

	err := fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// TODO: make sense of this scenario
		// skip any index files, since they are meant
		// to be read as dir content, later
		if d.Name() == "index.md" {
			return nil
		}

		node := &TreeNode{
			fs:         srcFS,
			SourcePath: path,
			Entry:      d,
			cache:      &nodeCache{},
			IsLeaf:     !d.IsDir(),
		}

		if path == "." {
			// if its ., set the root
			node.IsRoot = true
			treeRoot = node
		} else if path == d.Name() {
			// if path is equal to the name, start a new branch under root
			node.Parent = treeRoot
			currentBranch = node
		} else if d.IsDir() {
			// if its a dir, start a new branch under the current branch
			node.Parent = currentBranch
			currentBranch = node
		} else {
			// if its a file, add the parent without branching (leaf node)
			node.Parent = currentBranch
		}

		// if its not the root, add the root to the node
		// and add the node to the parents children
		if !node.IsRoot {
			node.Parent.Children = append(node.Parent.Children, node)
		}

		// always point to the tree root, even for the root itself
		node.Root = treeRoot

		return nil

	})

	if err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}

	return treeRoot, nil
}
