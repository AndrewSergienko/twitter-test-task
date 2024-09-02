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
COPY --from=builder /usr/local/bin/goose /usr/local/bin/goose
COPY --from=builder /app/migrations /app/migrations

CMD ["sh", "-c", "/app/goapp"]
CMD ["sh", "-c", "/usr/local/bin/goose -dir app/migrations postgres \"host=$DB_HOST port=$DB_PORT user=$COCKROACH_USER password=$COCKROACH_PASSWORD dbname=$COCKROACH_DATABASE\" up && /app/goapp"]
