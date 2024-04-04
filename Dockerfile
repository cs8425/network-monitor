# build:
# docker build -t network-monitor:latest .

FROM golang:alpine AS builder

RUN apk add --no-cache --update \
  ca-certificates \
  git

ENV CGO_ENABLED=0

Add . /src
WORKDIR /src

RUN go build -tags=release -ldflags="-w -s" -trimpath -o /tmp/network-monitor .



FROM scratch
WORKDIR /srv

# the tls certificates:
# NB: this pulls directly from the upstream image, which already has ca-certificates:
#COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# the program:
COPY --from=builder /tmp/network-monitor /

ENTRYPOINT ["/network-monitor"]
CMD ["-f", "/srv/target.txt"]

