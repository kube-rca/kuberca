FROM golang:1.22-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM scratch

COPY --from=builder /build/main .

ENTRYPOINT ["./main"]

EXPOSE 8082
