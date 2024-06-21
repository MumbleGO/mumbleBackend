
FROM golang:1.22.4 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /workspace/cmd/api
RUN CGO_ENABLED=0 go build -o /workspace/out/app

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /workspace/out/app /app/app

ENTRYPOINT ["/app/app"]

EXPOSE 3000

