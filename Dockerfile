FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/kfz ./main.go

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/kfz /app/kfz
COPY application.yaml /app/application.yaml
COPY templates /app/templates

ENV DEV=1 \
    TLS=false \
    TZ=Asia/Shanghai

EXPOSE 8888

CMD ["/app/kfz"]
