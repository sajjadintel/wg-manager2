.PHONY: fmt test vet install package
all: test vet install

LATEST_TAG_COMMIT = $(shell git rev-list --tags --max-count=1)
export VERSION = $(shell git describe --tags $$LATEST_TAG_COMMIT 2>/dev/null)

fmt:
	go fmt ./...

test:
	go test -short ./...

vet:
	go vet ./...

install:
	go install ./...

package:
	go build -ldflags "-X main.appVersion=$$VERSION" .
	nfpm --config packaging/nfpm.yaml pkg --target wireguard-manager.deb
