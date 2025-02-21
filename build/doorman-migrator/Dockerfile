#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.22-alpine AS builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

COPY . /go/src

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /go/bin/migrate-v1-v2 -trimpath -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/migrate-v1-v2/main.go
#####################################################################################
FROM gcr.io/distroless/static-debian12 AS runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="doorman" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      description="This image is used to migrate doorman from v1 to v2" 

WORKDIR /app

COPY --from=builder /go/bin/migrate-v1-v2 /app/
COPY build/doorman-migrator/config.yaml /app/

ENTRYPOINT ["./migrate-v1-v2"]
