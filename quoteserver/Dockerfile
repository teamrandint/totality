# build stage
FROM golang:alpine AS build-env
COPY . /go/src/seng468/quoteserver
RUN apk add --no-cache git \
    && go get github.com/patrickmn/go-cache \
    && go get github.com/shopspring/decimal \
    && cd /go/src/seng468/quoteserver \
    && go build -o quoteserver

# final stage
FROM alpine

ARG quoteaddr
ENV quoteaddr=$quoteaddr
ARG quoteport
ENV quoteport=$quoteport
ARG auditaddr
ENV auditaddr=$auditaddr
ARG auditport
ENV auditport=$auditport

ARG legacyquoteaddr
ENV legacyquoteaddr=$legacyquoteaddr
ARG legacyquoteport
ENV legacyquoteport=$legacyquoteport

WORKDIR /app
COPY --from=build-env /go/src/seng468/quoteserver/quoteserver /app/
EXPOSE 44459-44459
EXPOSE 4444
ENTRYPOINT ./quoteserver
