FROM golang:1.25-latest

WORKDIR /kaytrade-server

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o server ./cmd/kay_trade

CMD ["./server"]
