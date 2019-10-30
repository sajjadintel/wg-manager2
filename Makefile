.PHONY: fmt test vet install package
all: test vet install

fmt:
	go fmt ./...

test:
	go test -short ./...

vet:
	go vet ./...

install:
	go install ./...

package:
	docker run --rm -v $(PWD):/repo -v $(PWD)/build:/build mullvadvpn/go-packager@sha256:5ad286057e8143df673c6a01df89939822c93f3871c23cd85188a38af4e07d21
