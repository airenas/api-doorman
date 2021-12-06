FROM iron/base
LABEL Description="This image is used to host the API doorman service" \ 
      Maintainer="airenass@gmail.com" Vendor="Airenas Vaiciunas" Version="1.0" 

WORKDIR /app

RUN addgroup app && adduser -S -G app app

RUN mkdir -p /app && chown -R app:app /app 

COPY data/doorman /app/
COPY config.yaml /app/

RUN chown app:app /app/* /app
 
USER ap

ENTRYPOINT ["./doorman"]
