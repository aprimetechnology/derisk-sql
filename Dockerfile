# the official golang image works for building, but is ~1 GB in size
FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.sum /app/go.sum
COPY go.mod /app/go.mod
COPY cmd /app/cmd
COPY internal /app/internal
COPY pkg /app/pkg
COPY analyzers /app/analyzers

# ensure all the tests are passing before proceeding
RUN go test ./...

# go install ./... to ensure binaries are compiled for ALL analyzers/* directories
RUN go install ./...

# debian's stable-slim image:
# - has everything a compiled golang program expects (libc/linking wise)
# - has an incredibly tiny image size (~70 MB)
# - still has useful container debugging binaries (eg: bash, ls, etc)
FROM debian:stable-slim AS deploy-stage

WORKDIR /

RUN apt-get update && \
    # install root certificates so built binary can interact with HTTPS
    # install git for use as needed as stable-slim image does not have git
    apt-get install -y ca-certificates git

COPY --from=build-stage /go/bin/* /usr/local/bin
