FROM golang:1.23.3-alpine3.20

ARG PORT # injected from Railway at build time

ENV APP_HOME /go/src/twitchsongrequests

WORKDIR "${APP_HOME}"

COPY . "${APP_HOME}"

RUN go mod download

RUN go mod verify

RUN go build .

EXPOSE $PORT

CMD ["./twitchsongrequests"]