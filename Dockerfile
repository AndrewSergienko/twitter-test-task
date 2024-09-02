FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o goapp .

RUN curl -fsSL \
        https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
        sh

FROM alpine:latest

COPY --from=builder /app/goapp /app/goapp
COPY --from=builder /app/certs /app/certs

CMD ["sh", "-c", "/app/goapp"]
