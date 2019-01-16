FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
COPY ./codex-travis-testrunner /acceptance/src/codex-travis-testrunner
COPY ./common /go/src/common
WORKDIR /acceptance/src/codex-travis-testrunner

RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/codex-travis-testrunner


FROM scratch
COPY --from=builder /go/bin/codex-travis-testrunner /go/bin/codex-travis-testrunner
COPY ./features/helloworld /features
ENTRYPOINT ["/go/bin/codex-travis-testrunner","-feature=/features","-tags=HelloWorld001"]