.PHONY: all go clean

MAIN_SRC=cmd/main.go

go: run

clean:
	rm -f main linux docker-build

main: $(MAIN_SRC)
	go build -o main -v $(MAIN_SRC)
	@chmod +x main

linux: $(MAIN_SRC)
	GOOS=linux GOARCH=amd64 go build -o main -v $(MAIN_SRC)

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
