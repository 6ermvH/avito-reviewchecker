FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY configs ./configs
EXPOSE 8080
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux
ENTRYPOINT ["go", "run", "./cmd/reviewchecker"]

