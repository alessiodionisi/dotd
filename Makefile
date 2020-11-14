format:
	goimports -w .
.PHONY: format

lint:
	golangci-lint run ./...
.PHONY: lint

release:
	goreleaser --skip-publish --rm-dist
.PHONY: release

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist
.PHONY: snapshot
