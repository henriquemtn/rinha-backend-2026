FROM golang:1.22-alpine AS build

WORKDIR /app

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/. ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server .

FROM alpine:3.20

WORKDIR /app

COPY --from=build /bin/server /app/server
COPY resources /app/resources

ENV NORMALIZATION_PATH=/app/resources/normalization.json
ENV MCC_RISK_PATH=/app/resources/mcc_risk.json
ENV REFERENCES_PATH=/app/resources/references.json.gz

EXPOSE 9999

CMD ["/app/server"]
