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
	docker build . -t wireguard-manager
	docker run --rm -v $(PWD):/repo -v $(PWD)/build:/build wireguard-manager
