FROM golang:1.14.3-alpine3.11 AS BUILD

RUN apk add stress-ng build-base

WORKDIR /app

ADD /go.mod /app/
ADD /go.sum /app/

RUN go mod download

ADD / /app/
RUN go test ./...
# ./... -run ^TestProcessStatsBasic$
