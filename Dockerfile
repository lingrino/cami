FROM scratch
LABEL maintainer="sean@lingrino.com"
COPY cami /
ENTRYPOINT ["/cami"]
