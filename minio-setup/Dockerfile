FROM minio/mc:RELEASE.2024-04-18T16-45-29Z

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
