FROM golang:1.25-rc-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY internal internal
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-gateway .

FROM golang:1.25-rc-alpine AS runner

RUN adduser -D gatekeeper

COPY --from=builder /app/api-gateway /app/api-gateway
COPY --from=builder /app/internal/router/route_table.json /app/internal/router/route_table.json

RUN chown -R gatekeeper:gatekeeper /app
RUN chmod +x /app/api-gateway

WORKDIR /app

EXPOSE 8000

USER gatekeeper

CMD ["./api-gateway"]