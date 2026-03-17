FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /athena ./cmd/athena

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /athena /usr/local/bin/athena

EXPOSE 8080
ENTRYPOINT ["athena"]
CMD ["serve", "-f", "/data/faq.csv"]
