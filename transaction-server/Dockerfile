# build stage
FROM golang:alpine AS build-env
COPY . /go/src/seng468/transaction-server
RUN apk add --no-cache git \
    && go get github.com/garyburd/redigo/redis \
    && go get github.com/patrickmn/go-cache \
    && go get github.com/pkg/profile \
    && go get github.com/shopspring/decimal \
    && go get golang.org/x/sync/syncmap \
    && cd /go/src/seng468/transaction-server \
    && go build -o transactionserve

# final stage
FROM alpine

ARG transaddr
ENV transaddr=$transaddr
ARG transport
ENV transport=$transport
ARG dbaddr
ENV dbaddr=$dbaddr
ARG dbport
ENV dbport=$dbport
ARG auditaddr
ENV auditaddr=$auditaddr
ARG auditport
ENV auditport=$auditport
ARG quoteaddr
ENV quoteaddr=$quoteaddr
ARG quoteport
ENV quoteport=$quoteport
ARG triggeraddr
ENV triggeraddr=$triggeraddr
ARG triggerport
ENV triggerport=$triggerport

WORKDIR /app
COPY --from=build-env /go/src/seng468/transaction-server/transactionserve /app/
EXPOSE 44455-44459
ENTRYPOINT ./transactionserve 