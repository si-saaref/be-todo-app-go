FROM golang:1.24-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o binary

ENTRYPOINT [ "/app/binary" ]