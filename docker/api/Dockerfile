FROM golang:1.19-alpine as build

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/src/2fas

COPY go.mod go.sum ./

RUN go mod download -x

COPY . .

RUN mkdir -p bin

RUN go build -trimpath -o bin/api ./cmd/api
RUN go build -trimpath -o bin/migrate ./cmd/migrate/

FROM alpine:latest

RUN adduser 2fas -D

USER 2fas

WORKDIR /home/2fas/

COPY --from=build /go/src/2fas/bin/* /usr/local/bin/
COPY ./migrations /migrations
COPY ./config/config.yml ./config.yml
COPY ./docker/api/run_api.sh /usr/local/bin/run_api.sh

CMD ["run_api.sh"]