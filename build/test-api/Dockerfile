FROM iron/base
LABEL Description="This image is used to host the API doorman service" \ 
      Maintainer="airenass@gmail.com" Vendor="Airenas Vaiciunas" Version="1.0" 

WORKDIR /app

COPY data/test-api /app/

ENTRYPOINT ["./test-api"]
