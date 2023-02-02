# Also update GitHub Actions workflow when bumping
FROM docker.io/library/golang:1.20 AS builder

WORKDIR /src/
COPY . .
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
RUN govulncheck ./...
RUN CGO_ENABLED=0 go build -ldflags='-extldflags=-static' -o /bin/delic ./cmd

FROM docker.io/library/alpine:latest
COPY --from=builder /bin/delic /delic
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /
ENTRYPOINT ["./delic"]
