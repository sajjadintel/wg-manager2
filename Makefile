.PHONY: fmt test vet install nfpm package
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

nfpm:
	go build -ldflags "-X main.appVersion=$$VERSION" .
	mkdir -p build
	nfpm --config packaging/nfpm.yaml pkg --target ./build/wireguard-manager.deb

package:
	docker build . -t wireguard-manager
	docker run --rm -v $(PWD)/build:/wireguard-manager/build wireguard-manager
