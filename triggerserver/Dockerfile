# build stage
FROM golang:alpine AS build-env
COPY . /go/src/seng468/triggerserver
RUN apk add --no-cache git \
    && go get github.com/garyburd/redigo/redis \
    && go get github.com/shopspring/decimal \
    && go get golang.org/x/sync/syncmap \
    && cd /go/src/seng468/triggerserver \
    && go build -o triggerserver

# final stage
FROM alpine

ARG triggeraddr
ENV triggeraddr=$triggeraddr
ARG triggerport
ENV triggerport=$triggerport
ARG auditaddr
ENV auditaddr=$auditaddr
ARG auditport
ENV auditport=$auditport
ARG transaddr
ENV transaddr=$transaddr
ARG transport
ENV transport=$transport
ARG quoteaddr
ENV quoteaddr=$quoteaddr
ARG quoteport
ENV quoteport=$quoteport

WORKDIR /app
COPY --from=build-env /go/src/seng468/triggerserver/triggerserver /app/
EXPOSE 44455-44459
ENTRYPOINT ./triggerserver