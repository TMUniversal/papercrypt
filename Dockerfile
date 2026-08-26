FROM debian:bookworm-slim

LABEL repository="https://github.com/TMUniversal/papercrypt" \
    homepage="https://github.com/TMUniversal/papercrypt/blob/main/README.md" \
    maintainer="TMUniversal <me@tmuniversal.eu>" \
    Vendor="TMUniversal" \
    org.opencontainers.image.base.name="debian:bookworm-slim" \
    org.opencontainers.image.title="PaperCrypt" \
    org.opencontainers.image.description="PaperCrypt is a Go-based command-line tool designed to enhance the security of your sensitive data through the generation of printable backup documents." \
    org.opencontainers.image.authors="TMUniversal <me@tmuniversal.eu> (https://tmuniversal.eu)" \
    org.opencontainers.image.vendor="TMUniversal" \
    # Source is overridden by the build system
    org.opencontainers.image.source="https://github.com/TMUniversal/papercrypt" \
    org.opencontainers.image.documentation="https://github.com/TMUniversal/papercrypt/blob/main/README.md" \
    org.opencontainers.image.url="https://github.com/users/TMUniversal/packages/container/package/papercrypt" \
    org.label-schema.usage="docker run --rm -it -v \$PWD:/data ghcr.io/tmuniversal/papercrypt:latest generate -i /data/myfile.txt -o /data/output.pdf --purpose 'Backup' --comment 'This is a backup of myfile.txt'"

RUN apt-get update && \
    apt-get install -y --no-install-recommends libzxing3 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/papercrypt /usr/local/bin/papercrypt
ENTRYPOINT ["papercrypt"]
