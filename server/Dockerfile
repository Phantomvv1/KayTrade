FROM golang:1.25

WORKDIR /kaytrade-server

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o server ./cmd/kaytrade

CMD ["./server"]
