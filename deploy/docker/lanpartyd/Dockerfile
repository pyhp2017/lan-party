FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /lanpartyd ./cmd/lanpartyd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /lanpartyd /lanpartyd
ENTRYPOINT ["/lanpartyd"]
