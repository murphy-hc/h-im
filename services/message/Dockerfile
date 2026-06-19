FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY bin/server /app/bin/server
COPY configs /app/configs/

EXPOSE 9102
EXPOSE 8102

ENTRYPOINT ["/app/bin/server", "-conf", "/app/configs"]
