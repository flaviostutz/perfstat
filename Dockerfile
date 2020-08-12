FROM golang:1.14.3-alpine3.11 AS BUILD

RUN apk add stress-ng build-base

ENV RUN_TESTS 'false'

WORKDIR /app

ADD /go.mod /app/
ADD /go.sum /app/

RUN go mod download

ADD / /app/
RUN echo "TEST stats" && cd /app/stats && go test -v
RUN echo "TEST detectors" && cd /app/detectors && go test -v

WORKDIR /app/cli

RUN go build -o /bin/perfstat
RUN chmod +x /bin/perfstat

CMD [ "/app/startup.sh" ]

