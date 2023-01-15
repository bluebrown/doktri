package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bluebrown/doktri/internal/engine"
	"github.com/radovskyb/watcher"
	"github.com/urfave/cli/v2"
)

type DevServer struct {
	Source string
	Dist   string
	Theme  string
	Author string
	Port   int
	ngn    engine.Engine
}

func Serve(cCtx *cli.Context) error {
	s := &DevServer{
		Source: cCtx.Args().First(),
		Dist:   cCtx.String("dist"),
		Theme:  cCtx.String("theme"),
		Author: cCtx.String("author"),
		Port:   cCtx.Int("port"),
	}
	return s.Serve()
}

func (s *DevServer) Serve() error {
	s.makeEngine()
	if err := build(s.ngn); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	w := watcher.New()
	w.SetMaxEvents(1)
	errC := make(chan error)

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Printf("\nchange detected: %s\n", event.Path)
				s.makeEngine()
				if err := build(s.ngn); err != nil {
					fmt.Printf("render: %v\n", err)
				}
			case err := <-w.Error:
				errC <- fmt.Errorf("watch: %w", err)
				return
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(s.ngn.DocsDir()); err != nil {
		return fmt.Errorf("watch sources: %w", err)
	}

	if err := w.AddRecursive(s.ngn.ThemeTemplatesDir()); err != nil {
		return fmt.Errorf("watch theme templates: %w", err)
	}

	if err := w.AddRecursive(s.ngn.ThemeAssetsDir()); err != nil {
		return fmt.Errorf("watch theme assets: %w", err)
	}

	if err := w.AddRecursive(s.ngn.ThemeAssetsDir()); err != nil {
		return fmt.Errorf("watch extra assets: %w", err)
	}

	if err := w.Add(s.ngn.MetaPath()); err != nil {
		return fmt.Errorf("watch meta yaml: %w", err)
	}

	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			errC <- fmt.Errorf("start watch: %w", err)
		}
	}()

	go func() {
		w.Wait()
		fmt.Printf("\n- Serving content on http://localhost:%d 📚\n\n", s.Port)
		http.Handle("/", http.FileServer(http.Dir(s.ngn.DistDir())))
		if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", s.Port), nil); err != http.ErrServerClosed {
			errC <- fmt.Errorf("server: %w", err)
		}
	}()

	return <-errC
}

func (s *DevServer) makeEngine() {
	s.ngn = engine.New(
		engine.WithSource(s.Source),
		engine.WithDist(s.Dist),
		engine.WithTheme(s.Theme),
		engine.WithAuthor(s.Author),
	)
}
