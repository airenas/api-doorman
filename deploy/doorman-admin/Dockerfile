FROM iron/base:3.14.0

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

EXPOSE 8000

RUN addgroup app && adduser -S -G app app

WORKDIR /app
RUN mkdir -p /app && chown -R app:app /app 

COPY data/doorman-admin /app/
COPY config.yaml /app/
RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./doorman-admin"]
