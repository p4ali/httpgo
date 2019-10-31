# usage: make version=1.0.0 goos=darwin|linux|windows release

# always run these targets
.PHONY: all go clean vet lint release

# variables
MAIN_SRC=cmd/main.go
OUT := main
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
PKG := ./...
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

goos=darwin
version := $(shell git describe --always --long --dirty)

go: run

clean:
	rm -rf main docker-build bin release

main: $(MAIN_SRC) vet lint
	go build -i -v -o ${OUT} -ldflags="-X main.version=${version}" $(MAIN_SRC)
	@chmod +x main

# available goos: linux|darwin|windows
release: $(MAIN_SRC)
	@mkdir -p bin release
	GOOS=$(goos) GOARCH=amd64 go build -o bin/$(goos)/httpgo -v -ldflags="-extldflags \"-static\" -w -s -X main.version=${version}" $(MAIN_SRC)
	@tar cvfz release/httpgo-$(version)-$(goos).tar.gz -C bin/$(goos) httpgo
	@sha256sum release/httpgo-$(version)-$(goos).tar.gz > release/httpgo-$(version)-$(goos).tar.gz.sha256

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

run: main
	./main -port 8000 -name httpgo

test:
	go clean -testcache ./...
	go test -v ./httpgo/

docker-build: Dockerfile $(MAIN_SRC)
	docker build -t p4ali/httpgo:0.0.1 .
	@touch docker-build

docker-run: docker-build
	docker run -d --rm --name httpgo -p12345:12345 -it p4ali/httpgo:0.0.1 /bin/httpgo -port 12345
