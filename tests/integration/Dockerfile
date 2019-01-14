FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
COPY ./codex-testrunner $GOPATH/src/mypackage/codex-testrunner/
WORKDIR $GOPATH/src/mypackage/codex-testrunner/

RUN go get -d -v

RUN go build -o /go/bin/codex-testrunner


FROM scratch
COPY --from=builder /go/bin/codex-testrunner /go/bin/codex-testrunner
COPY ./features /features
ENTRYPOINT ["/go/bin/codex-testrunner","-feature=/features","-tags=HelloWorld001"]