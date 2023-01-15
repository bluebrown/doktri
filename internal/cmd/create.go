package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func Create(cCtx *cli.Context) error {
	title := cCtx.Args().First()
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	p := cCtx.String("dir")
	if p == "" {
		p = "."
	}

	return create(p, title)
}

func create(dir, title string) error {
	name := fmt.Sprintf("%s-%s.md",
		time.Now().Format("2006-01-02"),
		strings.ToLower(strings.ReplaceAll(title, " ", "-")))

	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("# %s\n", title))
	return err
}
