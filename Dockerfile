FROM golang:1.12.9 AS builder

RUN mkdir /build
COPY . /build/
WORKDIR /build 

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o relay .

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

FROM scratch

COPY --from=builder /build/  /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app

ENTRYPOINT ["./relay"]