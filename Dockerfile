FROM golang:1.22-alpine AS builder

COPY ./rpc-proxy  .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /rpc-proxy

FROM scratch AS runner

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /rpc-proxy /app/rpc-proxy

EXPOSE 8080

ENV PORT=8080

USER 65532:65532

CMD ["/app/rpc-proxy"]