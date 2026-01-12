FROM golang:1.25-latest

WORKDIR /kaytrade

COPY go.mod go.sum ./
RUN go mod download

ENV KAYTRADE_ENV=docker

COPY . .

RUN go build -o client ./cmd/kay_trade

CMD ["./client"]
