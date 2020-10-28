FROM iron/base
LABEL Description="This image is used to host the API doorman admin service" \ 
      Maintainer="airenass@gmail.com" Vendor="Airenas Vaiciunas" Version="1.0" 

WORKDIR /app

COPY data/doorman-admin /app/
COPY config.yaml /app/

ENTRYPOINT ["./doorman-admin"]
