#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM huacnlee/gobackup:v2.13.1

ENV POSTGRES_USER=postgres
ENV POSTGRES_DB=doorman
ENV POSTGRES_HOST=postgres
ENV POSTGRES_PORT=5432

ENV S3_BUCKET=doorman-db
ENV S3_KEEP=30
ENV S3_REGION=us-east-1
ENV S3_URL=http://minio:9000
ENV S3_PATH=backups
ENV S3_PATH_LONG=backups-long

ENV NOTIFY_TELEGRAM_ON_SUCCESS=false

# every 6 hours
ENV SCHEDULE=6h
# every 90 days
ENV SCHEDULE_LONG=2160h


COPY ./db/backup/gobackup.yml /etc/gobackup/gobackup.yml

ENTRYPOINT ["/usr/local/bin/gobackup"]
CMD ["run"]
