FROM golang:1.22.5-bookworm AS builder

WORKDIR /app

# We comment this to disable upx for faster builds for now
#ARG upx_version=4.2.4
ARG arch=arm64

COPY manager/go.mod manager/go.sum ./
RUN go mod download

ENV CGO_ENABLED=0
ENV GOARCH=arm64
ENV GOOS=linux

COPY manager/ .

# RUN go build -o rest-api -a -ldflags="-s -w" -installsuffix nocgo ./cmd/rest-api/
RUN go build -o rest-api ./cmd/rest-api/
RUN go build -o worker ./cmd/worker/

FROM scratch
WORKDIR /app
COPY --from=builder /app .
EXPOSE 8080
