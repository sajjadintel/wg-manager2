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
	docker run --rm -v $(PWD):/repo -v $(PWD)/build:/build mullvadvpn/go-packager@sha256:a54c9376a54d5b1a38710a11f86fc3a093272efffb64506b337b6f5d5b265d4d
