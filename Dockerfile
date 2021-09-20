# defaults which may be overridden from the build command
ARG GO_VERSION=1.16
ARG ALPINE_VERSION=3.12

# build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

COPY . /go/src/github.com/jabbors/gowold
WORKDIR /go/src/github.com/jabbors/gowold
#ARG APP_VERSION=0.0
#RUN go install -ldflags="-X \"main.version=${APP_VERSION}\""
# RUN go mod init
RUN go install

# final stage
FROM alpine:${ALPINE_VERSION}

RUN apk --no-cache add ca-certificates bash curl
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /go/bin/gowold /usr/bin/gowold
USER nobody:nobody
CMD [ "/usr/bin/gowold" ]
EXPOSE 8080
