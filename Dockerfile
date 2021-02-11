# Set the version of go we wish to use and compile with.
ARG go_version=1.15

# Setup our base image.
# This is used for compilation and running tests.
FROM golang:${go_version} as base

WORKDIR /app

COPY go.mod go.sum ./

RUN set -eux; \
    go mod download; \
    go get -u golang.org/x/lint/golint; \
    go get -u honnef.co/go/tools/cmd/staticcheck

COPY . .

ENTRYPOINT ["go"]

# Setup our compiler and build the production binaries.
FROM base as compiler

ARG go_proxy
ENV GOPROXY=${go_proxy}

COPY . .

RUN set -eux; \
    GOOS=linux CGO_ENABLED=0 GOGC=off GOARCH=amd64 go build -o api cmd/api/main.go; \
    GOOS=linux CGO_ENABLED=0 GOGC=off GOARCH=amd64 go build -o scraper cmd/scraper/main.go; \
    chmod +x api && chmod +x scraper

# Build an image containing certs ready to use in our alpine image.
FROM alpine as certs

RUN apk add -U --no-cache ca-certificates

# Download an ubuntu image for the sole purpose of creating a non-root user for scratch.
FROM alpine as userbuilder

ARG uid=10001
ARG gid=10001

RUN echo "scratchuser:x:${uid}:${gid}::/home/scratchuser:/bin/sh" > /scratchpasswd

# Build our scratch image with the production build binary from compiler.
FROM scratch as production-base

# Import certificates from the certs image.
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Import the binary from our compiler image.
COPY --from=compiler /app/api /
COPY --from=compiler /app/scraper /

# Import our user from our userbuilder image.
COPY --from=userbuilder /scratchpasswd /etc/passwd

USER scratchuser

EXPOSE 8901

ENV BINARY=api

FROM production-base as api

ENTRYPOINT ["/api"]

FROM production-base as scraper

ENTRYPOINT ["/scraper"]