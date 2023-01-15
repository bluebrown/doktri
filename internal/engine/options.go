package engine

type Options struct {
	source      string
	dist        string
	theme       string
	chromaStyle string
}

type Option func(opts *Options)

func WithAuthor(author string) Option {
	return func(opts *Options) {
		if author != "" {
			POST_AUTHOR = author
		}
	}
}

func WithContextPath(path string) Option {
	return func(opts *Options) {
		if path != "" {
			CONTEXT_PATH = path
		}
	}
}

func WithSource(src string) Option {
	return func(opts *Options) {
		opts.source = src
	}
}

func WithDist(dist string) Option {
	return func(opts *Options) {
		opts.dist = dist
	}
}

func WithTheme(theme string) Option {
	return func(opts *Options) {
		opts.theme = theme
	}
}

func WithChromaStyle(style string) Option {
	return func(opts *Options) {
		opts.chromaStyle = style
	}
}
