PKGS = $(shell go list ./...)

passd:
	mkdir -p build && go build -o ./bin/passd ./cmd/passd/
.PHONY: passd

test:
	go test -race -cover -covermode=atomic $(PKGS)
.PHONY: test
