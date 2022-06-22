#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.17-alpine as builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

# precache go.mod dependencies
COPY go.mod /go/src/
COPY go.sum /go/src/
RUN go mod download

# copy src
COPY . /go/src

RUN CGO_ENABLED=0 go build -o /go/bin/doorman-admin -trimpath -ldflags "-s -w -X main.version=$BUILD_VERSION" cmd/doorman-admin/main.go
#####################################################################################
FROM alpine:3.15 as runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="doorman-admin" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to host the API doorman admin service" \
      description="This image is used to host the API doorman admin service" 

WORKDIR /app

ENV UID=999
EXPOSE 8000

RUN addgroup app && adduser -S -G app app --uid $UID app

RUN mkdir -p /app && chown -R app:app /app   

COPY --from=builder /go/bin/doorman-admin /app/
COPY build/doorman-admin/config.yaml /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./doorman-admin"]
