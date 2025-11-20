FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o reviewchecker ./cmd/reviewchecker

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/reviewchecker .
COPY configs ./configs
ENV CONFIG_PATH=/app/configs/developer.yaml
EXPOSE 8080
ENTRYPOINT ["/app/reviewchecker"]
