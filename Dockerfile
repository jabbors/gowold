# defaults which may be overridden from the build command
ARG GO_VERSION=1.18
ARG ALPINE_VERSION=3.16

# build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

COPY . /go/src/github.com/jabbors/gowold
WORKDIR /go/src/github.com/jabbors/gowold
ARG APP_VERSION=0.2
RUN go install -ldflags="-X \"main.version=${APP_VERSION}\""

# final stage
FROM alpine:${ALPINE_VERSION}

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /go/bin/gowold /usr/bin/gowold
USER nobody:nobody
CMD [ "/usr/bin/gowold" ]
EXPOSE 8080
