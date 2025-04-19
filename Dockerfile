FROM golang:1.24.2-bookworm

WORKDIR /app/

COPY src/ /app/

RUN go build main.go


CMD ["/app/shakespeare"]