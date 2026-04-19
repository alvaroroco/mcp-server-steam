FROM golang:1.23.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /mcp-server-steam ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /mcp-server-steam .

ENV MCP_TRANSPORT=stdio
ENV MCP_PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/mcp-server-steam"]
