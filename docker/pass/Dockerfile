FROM golang:1.21-alpine as build

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/src/2fas

COPY go.mod go.sum ./

RUN go mod download -x

COPY . .

RUN mkdir -p bin

RUN go build -trimpath -o bin/pass ./cmd/pass/main.go

FROM alpine:latest

RUN adduser 2fas -D

USER 2fas

WORKDIR /home/2fas/

COPY --from=build /go/src/2fas/bin/* /usr/local/bin/

CMD ["pass"]
