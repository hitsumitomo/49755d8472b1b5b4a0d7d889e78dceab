FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

#RUN go mod tidy

EXPOSE 8080

RUN go build -o app .

FROM alpine:latest

COPY --from=builder /app/app .

CMD ["./app"]