FROM golang:1.14.3-alpine3.11 AS BUILD

WORKDIR /app

ADD /go.mod /app/
ADD /go.sum /app/

RUN go mod download

ADD /main.go /app/
RUN go build -o /bin/perfstat


FROM alpine:3.12.0

COPY --from=BUILD /bin/perfstat /bin/
ADD startup.sh /

CMD [ "/startup.sh" ]
