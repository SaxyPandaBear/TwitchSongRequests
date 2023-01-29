FROM golang:1.19.5-alpine3.17

ENV APP_HOME /go/src/twitchsongrequests

WORKDIR "${APP_HOME}"

COPY . "${APP_HOME}"

RUN go mod download

RUN go mod verify

RUN go build .

EXPOSE 8000

CMD ["./twitchsongrequests"]