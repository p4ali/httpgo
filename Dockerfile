FROM golang:1.13-alpine as builder

# on|off|auto. auto will turn on module based on the existence of go.mod
ENV GO111MODULE=auto

# httpgo version
ENV VERSION=1.0.0

ENV PROJECT httpgo
WORKDIR /go/src/$PROJECT

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main -v -ldflags="-extldflags \"-static\" -w -s -X main.version=${VERSION}" cmd/main.go

FROM alpine:3.9 as release
ENV PROJECT httpgo

# see https://github.com/gliderlabs/docker-alpine/issues/191#issuecomment-314148406
RUN echo 'https://dl-3.alpinelinux.org/alpine/v3.9/main' > /etc/apk/repositories
RUN apk add --no-cache strace && apk update && apk add bash curl iptables bind-tools netcat-openbsd tcpdump busybox busybox-extras

WORKDIR /bin/

COPY --from=builder /go/src/$PROJECT/main httpgo
COPY --from=builder /go/src/$PROJECT/run_docker_tests.sh .
EXPOSE 8000

CMD ["/bin/httpgo", "-port", "8000", "-name", "httpgo"]
