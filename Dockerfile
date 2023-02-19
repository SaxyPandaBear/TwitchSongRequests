FROM golang:1.20.1-alpine3.17

ARG PORT # injected from Railway at build time

ENV APP_HOME /go/src/twitchsongrequests

WORKDIR "${APP_HOME}"

COPY . "${APP_HOME}"

RUN go mod download

RUN go mod verify

RUN go build .

EXPOSE $PORT

CMD ["./twitchsongrequests"]