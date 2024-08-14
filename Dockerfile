FROM golang:1.21-bullseye AS builder

WORKDIR /app
COPY . .
RUN go build -o /go/bin/findtfvars ./cmd/findtfvars
RUN go build -o /go/bin/tfvars-to-args ./cmd/tfvars-to-args

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/bin/findtfvars /usr/local/bin/findtfvars
COPY --from=builder /go/bin/tfvars-to-args /usr/local/bin/tfvars-to-args
CMD ["bash"]
