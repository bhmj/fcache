ARG GOLANG_VERSION
ARG GOLANG_IMAGE
ARG TARGET_DISTR_TYPE
ARG TARGET_DISTR_VERSION
ARG GOOS
ARG GOARCH

FROM golang:${GOLANG_VERSION}-${GOLANG_IMAGE} AS builder
ARG LDFLAGS
WORKDIR /source
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY vendor/ ./vendor/
ARG GOOS
ARG GOARCH
RUN GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -trimpath -o bin/fcache ./cmd

FROM --platform=${GOOS}/${GOARCH} ${TARGET_DISTR_TYPE}:${TARGET_DISTR_VERSION} AS fcache
ARG USER
RUN adduser -Ds /bin/sh ${USER}
RUN apk update && apk add --no-cache bind-tools ca-certificates && update-ca-certificates
WORKDIR /app
# app
COPY --from=builder /source/bin/fcache .
# config
COPY config/ config/
# sql migrations
COPY sql/ migrations/

ENTRYPOINT ["./fcache"]
