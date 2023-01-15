SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

GORELEASER_VERSION ?= v1.14.1
GORELEASER_CONFIG ?= hack/goreleaser.yaml

.PHONY: release-dryrun
release-dryrun: bin/goreleaser tag
	@bin/goreleaser release -f $(GORELEASER_CONFIG) --rm-dist --skip-publish

.PHONY: release
release: bin/goreleaser tag
	@bin/goreleaser release -f $(GORELEASER_CONFIG) --rm-dist

.PHONY: tag
tag:
	@$(eval git_tag=v$(shell docker run -v "$(CURDIR):/tmp" --workdir /tmp \
		--rm -u '$(shell id -u):$(shell id -g)' convco/convco version --bump))
	@git tag -m "bump" -f "$(git_tag)"

bin/goreleaser: bin
	@curl -fsSL https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_Linux_x86_64.tar.gz \
		| tar -C bin -xzf - goreleaser

bin:
	@mkdir -p bin
