FROM golang:1.23-bullseye AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y \
    gcc \
    libc-dev \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1

RUN go build -ldflags="-extldflags=-lpthread" -o /goodblast .

FROM gcr.io/distroless/base-debian11

COPY --from=builder /goodblast /goodblast

EXPOSE 8080

CMD ["/goodblast"]