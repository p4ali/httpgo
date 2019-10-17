.PHONY: all go clean

MAIN_SRC=cmd/main.go

go: run

clean:
	rm -f main docker-build

main: $(MAIN_SRC)
	go build -o main -v $(MAIN_SRC)
	@chmod +x main

run: main
	./main -port 8000 -name httpgo -version 0.0.1

test:
	go clean -testcache ./...
	go test ./httpgo/

docker-build: Dockerfile $(MAIN_SRC)
	docker build -t p4ali/httpgo:0.0.1 .
	@touch docker-build

docker-run: docker-build
	docker run -d --rm --name httpgo -p12345:12345 -it p4ali/httpgo:0.0.1 /bin/httpgo -port 12345 -version 0.0.1
