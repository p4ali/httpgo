# usage: make version=1.0.0 goos=darwin|linux|windows release

# always run these targets
.PHONY: all go clean release

# variables
MAIN_SRC=cmd/main.go
goos=darwin
version=1.0.0

go: run

clean:
	rm -rf main docker-build bin release

main: $(MAIN_SRC)
	go build -o main -v $(MAIN_SRC)
	@chmod +x main

# available goos: linux|darwin|windows
release: $(MAIN_SRC)
	@mkdir -p bin release
	GOOS=$(goos) GOARCH=amd64 go build -o bin/$(goos)/httpgo -v $(MAIN_SRC)
	@tar cvfz release/httpgo-$(version)-$(goos).tar.gz -C bin/$(goos) httpgo
	@sha256sum release/httpgo-$(version)-$(goos).tar.gz > release/httpgo-$(version)-$(goos).tar.gz.sha256

run: main
	./main -port 8000 -name httpgo -version 0.0.1

test:
	go clean -testcache ./...
	go test -v ./httpgo/

docker-build: Dockerfile $(MAIN_SRC)
	docker build -t p4ali/httpgo:0.0.1 .
	@touch docker-build

docker-run: docker-build
	docker run -d --rm --name httpgo -p12345:12345 -it p4ali/httpgo:0.0.1 /bin/httpgo -port 12345 -version 0.0.1
