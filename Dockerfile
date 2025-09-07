FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bili-parse-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bili-parse-api .
RUN mkdir -p data logs

EXPOSE 8080
CMD ["./bili-parse-api"]
