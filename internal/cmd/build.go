package cmd

import (
	"fmt"

	"github.com/bluebrown/doktri/internal/engine"
	"github.com/urfave/cli/v2"
)

func Build(cCtx *cli.Context) error {
	e := engine.New(
		engine.WithSource(cCtx.Args().First()),
		engine.WithDist(cCtx.String("dist")),
		engine.WithTheme(cCtx.String("theme")),
		engine.WithAuthor(cCtx.String("author")),
		engine.WithContextPath(cCtx.String("context")),
	)
	return build(e)
}

func build(e engine.Engine) error {
	fmt.Printf("\n- building content 🏗️\n")
	if err := e.Run(); err != nil {
		fmt.Printf("\n- Failure ❌\n")
		return err
	}
	fmt.Printf("\n- Done 👌\n")
	return nil
}
