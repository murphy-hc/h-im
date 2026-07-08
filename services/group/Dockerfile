FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

RUN addgroup --system app && adduser --system --ingroup app app
USER app

COPY bin/server /app/bin/server
COPY configs/ /app/configs/

EXPOSE 9104
EXPOSE 8104

ENTRYPOINT ["/app/bin/server", "-conf", "/app/configs"]
