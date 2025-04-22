FROM golang:1.24.2-bookworm

WORKDIR /app/

COPY src/ /app/

RUN go build -o main main.go


CMD ["/app/main"]