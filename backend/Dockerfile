FROM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/main .

ENTRYPOINT ["./main"]
EXPOSE 8080

LABEL org.opencontainers.image.source="https://github.com/kube-rca/kuberca"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.description="KubeRCA Backend"
LABEL org.opencontainers.image.documentation="https://github.com/kube-rca/kuberca/tree/main/backend"
LABEL org.opencontainers.image.vendor="KubeRCA"
LABEL org.opencontainers.image.title="KubeRCA backend"
