FROM golang:1.20-alpine as builder
RUN apk update && apk upgrade && \
  apk add --no-cache git openssh gcc libc-dev

WORKDIR /go/src/github.com/brianbroderick/lantern
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . . 

RUN go install -v /go/src/github.com/brianbroderick/lantern/

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /go/bin/lantern .

CMD ["./lantern"]
