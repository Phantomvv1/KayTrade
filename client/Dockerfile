FROM golang:1.25

WORKDIR /kaytrade

COPY go.mod go.sum ./
RUN go mod tidy

ENV KAYTRADE_ENV=docker

COPY . .

RUN go build -o client ./cmd/kaytrade

CMD ["./client"]
