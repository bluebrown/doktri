package main

import (
	"fmt"
	"os"

	"github.com/bluebrown/doktri/internal/cmd"
	"github.com/urfave/cli/v2"
)

var (
	version = "unknown"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf(`{"version": %q, "revision": %q, "date": %q, "buildBy": %q}`+"\n",
			cCtx.App.Version, commit, date, builtBy)
	}

	app := &cli.App{
		Name:    "doktri",
		Version: version,
		Usage:   "a static site generator",
		Commands: []*cli.Command{
			{
				Name:      "build",
				Aliases:   []string{"b"},
				Usage:     "build the static html content",
				ArgsUsage: "[src-dir]",
				Action:    cmd.Build,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "dist",
						Usage:       "output directory",
						DefaultText: "<src>/dist",
					},
					&cli.StringFlag{
						Name:        "theme",
						Usage:       "the theme to use",
						DefaultText: "<src>/.theme",
					},
					&cli.StringFlag{
						Name:    "author",
						Usage:   "global post author",
						EnvVars: []string{"DOKTRI_AUTHOR"},
					},
					&cli.StringFlag{
						Name:    "context",
						Usage:   "context path used when generating links",
						EnvVars: []string{"DOKTRI_CONTEXT"},
					},
					&cli.StringFlag{
						Name:    "chroma-style",
						Usage:   "chroma style to use for syntax highlighting",
						EnvVars: []string{"DOKTRI_CHROMA_STYLE"},
					},
				},
			},
			{
				Name:      "serve",
				Aliases:   []string{"s"},
				Usage:     "build and serve the static html content, with hot reload",
				ArgsUsage: "[src-dir]",
				Action:    cmd.Serve,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "dist",
						Usage:       "output directory",
						DefaultText: "<src>/dist",
					},
					&cli.StringFlag{
						Name:        "theme",
						Usage:       "the theme to use",
						DefaultText: "<src>/.theme",
					},
					&cli.StringFlag{
						Name:    "author",
						Usage:   "global post author",
						EnvVars: []string{"DOKTRI_AUTHOR"},
					},
					&cli.IntFlag{
						Name:  "port",
						Usage: "the port to listen on",
						Value: 3000,
					},
					&cli.StringFlag{
						Name:    "chroma-style",
						Usage:   "chroma style to use for syntax highlighting",
						EnvVars: []string{"DOKTRI_CHROMA_STYLE"},
					},
				},
			},
			{
				Name:      "init",
				Aliases:   []string{"i"},
				Usage:     "initialize a new project",
				ArgsUsage: "[target-dir]",
				Action:    cmd.InitProject,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "title",
						Aliases: []string{"t"},
						Usage:   "title to set in the meta.yaml",
						Value:   "My Page",
					},
				},
			},
			{
				Name:      "create",
				Usage:     "create a new post",
				ArgsUsage: "[title]",
				Aliases:   []string{"c"},
				Action:    cmd.Create,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Usage:   "directory to create the post in",
						Value:   ".",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
