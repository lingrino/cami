FROM scratch

# https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.ref.name="cami" \
    org.opencontainers.image.ref.title="cami" \
    org.opencontainers.image.description="A CLI for cleaning up unused AWS AMIs" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.authors="sean@lingrino.com" \
    org.opencontainers.image.url="https://lingrino.com" \
    org.opencontainers.image.documentation="https://lingrino.com" \
    org.opencontainers.image.source="https://github.com/lingrino/cami"

COPY cami /
ENTRYPOINT ["/cami"]
