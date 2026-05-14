FROM golang:1.24-alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p dataset && cp resources/references.json.gz resources/ && cp resources/normalization.json resources/ && cp resources/mcc_risk.json resources/
RUN CGO_ENABLED=1 go build -o indexer ./cmd/indexer
RUN CGO_ENABLED=1 go build -o api ./cmd/api
RUN ./indexer

FROM alpine:3.20
WORKDIR /app/dataset
COPY --from=builder /src/api /app/
COPY --from=builder /src/resources/ivf.bin .
COPY --from=builder /src/resources/normalization.json .
COPY --from=builder /src/resources/mcc_risk.json .
WORKDIR /app
EXPOSE 8080
CMD ["./api"]
