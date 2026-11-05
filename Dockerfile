FROM golang:1.23-bookworm AS builder

WORKDIR /app
COPY . .
RUN go build -o /go/bin/tfv ./cmd/tfv

FROM golang:1.23-bookworm
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/bin/tfv /usr/local/bin/tfv
CMD ["bash"]
