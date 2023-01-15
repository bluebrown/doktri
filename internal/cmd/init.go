package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bluebrown/doktri/internal/fsys"
	"github.com/urfave/cli/v2"
	"sigs.k8s.io/yaml"
)

func InitProject(cCtx *cli.Context) error {
	title := cCtx.String("title")
	outDir := cCtx.Args().First()
	if outDir == "" {
		outDir = "."
	}
	return initProject(title, outDir)
}

type social struct {
	Title  string `json:"title,omitempty"`
	Anchor string `json:"anchor,omitempty"`
	Icon   string `json:"icon,omitempty"`
}

type defaultMeta struct {
	Title   string   `json:"title,omitempty"`
	Socials []social `json:"socials,omitempty"`
}

func initProject(title, outDir string) error {

	// check if dir is empty
	ok, err := fsys.PathExists(outDir)
	if err != nil {
		return err
	}
	if ok {
		ok, err := fsys.IsEmptyDir(outDir)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("target dir must be empty")
		}
	}

	docsDir := filepath.Join(outDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("create docs dir at %q: %w", docsDir, err)
	}

	themeDir := filepath.Join(outDir, ".theme")
	if err := exec.Command("git", "clone", "https://github.com/bluebrown/doktri-theme-default", themeDir).Start(); err != nil {
		return fmt.Errorf("clone theme to %q: %w", themeDir, err)
	}

	if err := os.RemoveAll(filepath.Join(outDir, ".theme", ".git")); err != nil {
		return fmt.Errorf("remove theme .git: %w", err)
	}

	f, err := os.Create(filepath.Join(outDir, "meta.yaml"))
	if err != nil {
		return fmt.Errorf("create meta.yaml: %w", err)
	}

	defer f.Close()

	b, err := yaml.Marshal(&defaultMeta{
		Title: title,
		Socials: []social{
			{Title: "bluebrown", Anchor: "https://github.com/bluebrown/", Icon: "github"},
		},
	})

	if err != nil {
		return fmt.Errorf("marshal default meta: %w", err)
	}

	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("write default meta: %w", err)
	}

	f2, err := os.Create(filepath.Join(outDir, ".gitignore"))
	if err != nil {
		return fmt.Errorf("create .gitignore: %w", err)
	}

	defer f2.Close()

	if _, err := f2.WriteString("/dist/\n"); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}

	return create(docsDir, "My First Post")
}
